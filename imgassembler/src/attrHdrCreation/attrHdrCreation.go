/*
 * Copyright (c) 2020-2022, MaxLinear, Inc.
 * 
 * APIs to create attribute header and image with header
 *
 */
package attrHdrCreation

import (
	"encoding/json"
	"fmt"
	"strings"
	"io/ioutil"
	"os"
	"encoding/binary"
	"os/exec"
	"bytes"
	"strconv"
	"path/filepath"
	"path"
	"imgassembler/types"
)
func updateAttrElement (ImgAttrFromJson types.ImgAttr, attr types.AttributeElement) types.AttributeElement {
	if len(ImgAttrFromJson.TypeId) <= 0 {
		attr.ElementType = 0
	} else {
		attr.ElementType = hex2uint(ImgAttrFromJson.TypeId)
	}
	if len(ImgAttrFromJson.AttrVal) <= 0 {
		attr.ElementValue = 0
	} else {
		attr.ElementValue = hex2uint(ImgAttrFromJson.AttrVal)
	}
	return attr;
}
func hex2uint(hexStr string) uint32 {
	tmpstr := strings.Replace(hexStr, "0x", "", -1)
	result, err := strconv.ParseUint(tmpstr, 16, 64)
	if err != nil {
		fmt.Println(err)
		return 0;
	}
	return uint32(result)
 }
func createImgwithAttrHdr (imgFilePath string) error {
	var stderr bytes.Buffer
	//align image for 16 bytes
	tmppath := imgFilePath + ".tmp"
	opath := "of=" + imgFilePath + ".tmp"
	iptah := "if=" + imgFilePath
	cmd := exec.Command("/bin/dd", iptah, opath, "bs=16", "conv=sync")
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("createImgwithAttrHdr : padding image for alignment failed: %w %s", err, stderr.String())
	}

	dirName, fileName := path.Split(imgFilePath)

	newFileName := dirName + fileName + ".with_attr_hdr"
	args := "cat " + "attrhdr.bin " + imgFilePath + ".tmp" + " > " + newFileName
	fmt.Println(args)
	cmd = exec.Command("/bin/sh","-c", args)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("createImgwithAttrHdr : Creating image with attribute header failed : %w", err)
	}
	os.Remove("attrhdr.bin")
	os.Remove(tmppath)
	if strings.Contains(imgFilePath, ".trim") {
		os.Remove(imgFilePath)
	}
	return nil
}
func AttrHdr (imgAttrFilePath string , imgFilePath string , recovery int) error {

	var rbe int
	var imagesAttrs types.ImgsAttrs
	var attrsElement types.AttributeElement
	jFile, err := os.Open(imgAttrFilePath)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer jFile.Close()
	imagesArray, _ := ioutil.ReadAll(jFile)
	json.Unmarshal(imagesArray, &imagesAttrs)
	/*Adding attribute version TypeId 1 and Value 0*/
	attrsHdr := []types.AttributeElement{types.AttributeElement{1, 0}}
	for i := 0; i < len(imagesAttrs.ImagesArray); i++ {
		fmt.Println(imagesAttrs.ImagesArray[i].TypeId)
		fmt.Println(imagesAttrs.ImagesArray[i].AttrVal)
		if imagesAttrs.ImagesArray[i].TypeId == "0x80000009" {
			if imagesAttrs.ImagesArray[i].AttrVal == "0x00000001" {
				rbe = 1
			}
		}
		attrsElement = updateAttrElement(imagesAttrs.ImagesArray[i],
						attrsElement)
		attrsHdr = append(attrsHdr,attrsElement)
	}
	/*Adding Attribute TypeId 0 and value 0*/
	attrsHdr = append(attrsHdr,types.AttributeElement{0 , 0})
	fp, err := os.Create("attrhdr.bin")
	if err != nil {
		return fmt.Errorf("createHdrFile : Creating attribute header file failed : %w", err)
	}
	err = binary.Write(fp, binary.BigEndian, attrsHdr)
	if err != nil {
		return fmt.Errorf("createHdrFile : Creating attribute header failed : %w", err)
		fmt.Println(err)
	}
	fp.Close()
	const atthdrsz = 128
	hdrlen := "seek=128"
	cmd := exec.Command("/bin/dd","if=/dev/null","of=attrhdr.bin","obs=1",hdrlen)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("createHdrFile : padding attribute header file failed : %w", err)
	}

	//stripping the mkimage header if present
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("AttrHdr : Getting working directory failed: %w", err)
	}
	curdir := filepath.Dir(exe)
	path := curdir + "/mkimage"
	tmppath := imgFilePath
	if rbe != 1 {
		out, err := exec.Command(path, "-l", imgFilePath).Output()
		if (( err == nil ) && ( len(string(out)) > 0 ) &&
			strings.Contains(string(out), "Type") ) {
			fmt.Println("CreateUpTotalImg : mkimage header is present , removing it")
			tmppath = imgFilePath + ".trim"
			opath := "of=" + imgFilePath + ".trim"
			iptah := "if=" + imgFilePath
			cmd := exec.Command("/bin/dd", iptah, opath, "bs=1", "skip=64")
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			err = cmd.Run()
			if err != nil {
				return fmt.Errorf("AttrHdr : Removing mkimage header failed: %w", err)
			}
		}
	} else if recovery != 1 {
		fmt.Println("Handling RBE Image")
		args := "/usr/bin/xxd -p -l4 -s 0x844 " + imgFilePath + " | " + " /usr/bin/xxd -r -p > fw_temp.bin"
		cmd = exec.Command("/bin/sh","-c", args)
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			fmt.Println(err)
			return err
		}
		args = "/usr/bin/xxd -e -g4 fw_temp.bin " + "|" + " /usr/bin/xxd -r > fw_temp1.bin"
		cmd = exec.Command("/bin/sh","-c", args)
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			fmt.Println(err)
			return err
		}
		out, err := exec.Command("/usr/bin/xxd", "-p", "-l4", "fw_temp1.bin").Output()
		if err != nil {
			fmt.Println(err)
			return err
		}
		rbesize, err := strconv.ParseInt(strings.TrimSuffix(string(out), "\n"), 16, 64)
		fmt.Println(rbesize)
		fmt.Println(string(out))
		if err != nil {
			fmt.Println(err)
			return err;
		}
		tmppath = imgFilePath + ".trim"
		opath := "of=" + imgFilePath + ".trim"
		iptah := "if=" + imgFilePath
		count := "count=" + strconv.FormatInt(rbesize, 10)
		cmd = exec.Command("/bin/dd", iptah, opath, "bs=1", "skip=2124", count)
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			fmt.Println("Sign : Removing header from rbe image failed" + stderr.String())
			return err
		}
		os.Remove("fw_temp.bin")
		os.Remove("fw_temp1.bin")
	} else {
		fmt.Println("recovery image");
	}

	err = createImgwithAttrHdr ( tmppath )
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
