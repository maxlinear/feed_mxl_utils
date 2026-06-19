/*
 * Copyright (c) 2020-2022, MaxLinear, Inc.
 * 
 * APIs to assemble secureboot image
 *
 */
package sbImageCreation

import (
	"fmt"
	"os"
	"encoding/binary"
	"os/exec"
	"bytes"
	"path/filepath"
	"io/ioutil"
	"encoding/json"
	"imgassembler/types"
)

func Reconstruct_rbe(origImgPath string , signedimgpath string) error {
	var stderr bytes.Buffer
	const (
		rbe_addr = "0xBFF20000"
		socmode = "2"
	)

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("Reconstruct_rbe : Getting working directory failed: %w", err)
	}
	curdir := filepath.Dir(exe)
	/*create tail*/
	swapbinpath := curdir + "/swap_bin.pl"
	mktailpath := curdir + "/mk_intel_tail.pl"
	pad2align := curdir + "/pad2align.sh"
	cmd := exec.Command( swapbinpath, signedimgpath,"rbe_temp.bin")
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Reconstruct_rbe : swap_bin.pl failed while creating rbe tail: %w %s",
		                   err, stderr.String())
	}
	cmd = exec.Command( mktailpath ,"-i","rbe_temp.bin","-o","rbe_temp1.bin","-j",rbe_addr,"-m",socmode)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Reconstruct_rbe : mk_intel_tail.pl failed while creating rbe tail: %w %s",
		                   err, stderr.String())
	}
	outfile_tail := signedimgpath + ".tail"
	outfile_head := signedimgpath + ".head"

	cmd = exec.Command(swapbinpath, "rbe_temp1.bin", outfile_tail)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Reconstruct_rbe : swap_bin.pl failed while creating rbe tail: %w %s",
		                  err, stderr.String())
	}
	/*create head*/
	opath := "of=" + outfile_head
	iptah := "if=" + origImgPath
	cmd = exec.Command("/bin/dd", iptah, opath, "bs=1", "skip=64", "count=2048")
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Reconstruct_rbe : padding failed while creating rbe head: %w %s", err,
		                  stderr.String())
	}
	/*final rbe image*/
	cmd = exec.Command("/bin/cp" ,outfile_head,"rbe_temp2.bin")
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return err;
	}
	args := "cat " + outfile_tail + " >> rbe_temp2.bin"
	cmd = exec.Command("/bin/sh","-c", args)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Reconstruct_rbe : Creating rbe final image failed : %w", err)
	}
	/*padding to rbe image*/
	cmd = exec.Command( pad2align ,"-n","204800","rbe_temp2.bin")
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Reconstruct_rbe : pad2align.sh failed to pad rbe image: %w %s",
		                  err, stderr.String())
	}
	path := curdir + "/mkimage_2016"
	cmd = exec.Command(path,"-A","x86","-C","none","-T",
	                   "firmware","-a",rbe_addr,"-e",rbe_addr,"-n","RBE",
	                   "-d","rbe_temp2.bin",signedimgpath)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Reconstruct_rbe : adding header to rbe image failed: %w %s",
		                  err, stderr.String())
	}
	os.Remove("rbe_temp.bin")
	os.Remove("rbe_temp1.bin")
	os.Remove("rbe_temp2.bin")
	os.Remove(outfile_tail)
	os.Remove(outfile_head)
	return nil
}

func Reconstruct_recovery(signedimgpath string) error {
	var stderr bytes.Buffer
	const (
		recovery_addr = "0xBFF20000"
		socmode = "2"
	)
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("Reconstruct_recovery : Getting working directory failed: %w", err)
	}
	curdir := filepath.Dir(exe)
	/*create tail*/
	swapbinpath := curdir + "/swap_bin.pl"
	mktailpath := curdir + "/mk_intel_tail.pl"
	pad2align := curdir + "/pad2align.sh"
	cmd := exec.Command( swapbinpath ,signedimgpath,"recovery_temp.bin")
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Reconstruct_recovery : swap_bin.pl failed while creating recovery image: %w %s",
		                  err, stderr.String())
	}
	//Generating BSTRAP (DFU) Download Image
	cmd = exec.Command( pad2align ,"-n","4","recovery_temp.bin")
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Reconstruct_recovery : pad2align.sh failed to pad recovery image: %w %s",
		                  err, stderr.String())
	}
	cmd = exec.Command( mktailpath ,"-i","recovery_temp.bin","-o","recovery_temp2.bin","-j",recovery_addr,"-m",socmode)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Reconstruct_recovery : mk_intel_tail.pl failed while creating rbe tail: %w %s",
		                  err , stderr.String())
	}
	cmd = exec.Command( swapbinpath ,"recovery_temp2.bin",signedimgpath )
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Reconstruct_recovery : swap_bin.pl failed while creating recovery image: %w %s",
		                  err, stderr.String())
	}
	//Generating ICCP (UART) Download Image ...
	cmd = exec.Command( "/usr/bin/touch" ,"dummy")
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Reconstruct_recovery : creating dummy file failed: %w", err)
	}
	ascfile := signedimgpath + ".asc"
	uartbintoasc := curdir + "/uart_bin2asc.pl"
	cmd = exec.Command( uartbintoasc ,"dummy","recovery_temp.bin",recovery_addr,"dummy","0",ascfile,socmode)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Reconstruct_recovery : recovery asc file creation failed: %w %s",
		                  err ,stderr.String())
	}
	os.Remove("recovery_temp.bin")
	os.Remove("recovery_temp2.bin")
	return nil

}

func AddHeader(attrFile string, outputpath string, origImgPath string, recovery int) error {

	var stderr bytes.Buffer
	var cmd *exec.Cmd
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("AddHeader : Getting working directory failed: %w", err)
	}
	curdir := filepath.Dir(exe)
	path := curdir + "/mkimage"
	uboot_mkimage := curdir + "/mkimage_2016"

	var imagesAttrs types.ImgsAttrs
	jFile, err := os.Open(attrFile)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer jFile.Close()
	imagesArray, _ := ioutil.ReadAll(jFile)
	json.Unmarshal(imagesArray, &imagesAttrs)
	for i := 0; i < len(imagesAttrs.ImagesArray); i++ {
		//fmt.Println(imagesAttrs.ImagesArray[i].TypeId)
		//fmt.Println(imagesAttrs.ImagesArray[i].AttrVal)
		if imagesAttrs.ImagesArray[i].TypeId == "0x80000009" {
			if imagesAttrs.ImagesArray[i].AttrVal == "0x00000001" {
				if recovery == 1 {
					fmt.Println( "recovery Image")
					err = Reconstruct_recovery(outputpath)
					if err != nil {
						return fmt.Errorf("AddHeader : recovery image creation failed: %w %s", err, stderr.String())
					}
				} else {
					fmt.Println( "RBE Image")
					err = Reconstruct_rbe(origImgPath , outputpath)
					if err != nil {
						return fmt.Errorf("AddHeader : rbe image creation failed: %w %s", err, stderr.String())
					}
				}
			}
			if imagesAttrs.ImagesArray[i].AttrVal == "0x00000002" {
				fmt.Println( "TEP firmware Image")
				outputmkimge := outputpath + ".mkimg"
				tmppath := outputpath + ".tmp"
				opath := "of=" + outputpath + ".tmp"
				iptah := "if=" + outputpath
				cmd = exec.Command("/bin/dd", iptah, opath, "bs=16", "conv=sync")
				cmd.Stderr = &stderr
				err = cmd.Run()
				if err != nil {
					return fmt.Errorf("AddHeader : Padding Tepfirmware image failed: %w %s", err, stderr.String())
				}
				cmd = exec.Command(path,"-A","x86","-C","none","-T",
				"firmware","-n","TEP firmware",
				"-d",tmppath,outputmkimge)
				cmd.Stderr = &stderr
				err = cmd.Run()
				if err != nil {
					return fmt.Errorf("AddHeader : adding header to TEP firmware failed image failed: %w %s",
					                  err, stderr.String())
				}
				os.Remove(tmppath)
				os.Remove(outputpath)
				os.Rename(outputmkimge , outputpath);

			}
			if imagesAttrs.ImagesArray[i].AttrVal == "0x00000003" {
				fmt.Println( "uboot Image")
				outputmkimge := outputpath + ".mkimg"
				tmppath := outputpath + ".tmp"
				opath := "of=" + outputpath + ".tmp"
				iptah := "if=" + outputpath
				cmd = exec.Command("/bin/dd", iptah, opath, "bs=16", "conv=sync")
				cmd.Stderr = &stderr
				err = cmd.Run()
				if err != nil {
					return fmt.Errorf("AddHeader : Padding uboot image failed: %w %s", err, stderr.String())
				}
				cmd = exec.Command(uboot_mkimage,"-A","x86","-C","lzma","-T",
				"uboot","-a","0x08200000","-e","0x08200000","-n","u-boot image",
				"-d",tmppath,outputmkimge)
				cmd.Stderr = &stderr
				err = cmd.Run()
				if err != nil {
					return fmt.Errorf("AddHeader : adding header to uboot image failed: %w %s",
					                  err, stderr.String())
				}
				os.Remove(tmppath)
				os.Remove(outputpath)
				os.Rename(outputmkimge , outputpath);

			}
			if imagesAttrs.ImagesArray[i].AttrVal == "0x00000004" {
				fmt.Println( "Kernel Image")
				outputmkimge := outputpath + ".mkimg"
				tmppath := outputpath + ".tmp"
				opath := "of=" + outputpath + ".tmp"
				iptah := "if=" + outputpath
				cmd = exec.Command("/bin/dd", iptah, opath, "bs=16", "conv=sync")
				cmd.Stderr = &stderr
				err = cmd.Run()
				if err != nil {
					return fmt.Errorf("AddHeader : Padding kernel image failed: %w %s", err, stderr.String())
				}
				cmd = exec.Command(path,"-A","x86_64","-O","linux","-C","gzip","-T",
				"kernel","-a","0x2000000","-e","0x2000000","-n","X86_64 LEDE Linux-4.19",
				"-d",tmppath,outputmkimge)
				cmd.Stderr = &stderr
				err = cmd.Run()
				if err != nil {
					return fmt.Errorf("AddHeader : adding header to kernel image failed: %w %s",
					                  err, stderr.String())
				}
				os.Remove(tmppath)
				os.Remove(outputpath)
				os.Rename(outputmkimge , outputpath);

			}
			if imagesAttrs.ImagesArray[i].AttrVal == "0x00000005" {
				fmt.Println( "rootfs Image")
				outputmkimge := outputpath + ".mkimg"
				tmppath := outputpath + ".tmp"
				opath := "of=" + outputpath + ".tmp"
				iptah := "if=" + outputpath
				cmd = exec.Command("/bin/dd", iptah, opath, "bs=16", "conv=sync")
				cmd.Stderr = &stderr
				err = cmd.Run()
				if err != nil {
					return fmt.Errorf("AddHeader : Padding rootfs image failed: %w %s", err, stderr.String())
				}
				cmd = exec.Command(path,"-A","x86","-O","linux","-C","lzma","-T",
				"filesystem","-a","0x00","-e","0x00","-n","LEDE RootFS",
				"-d",tmppath,outputmkimge)
				cmd.Stderr = &stderr
				err = cmd.Run()
				if err != nil {
					return fmt.Errorf("AddHeader : adding header to rootfs image failed: %w %s",
					                  err, stderr.String())
				}
				os.Remove(tmppath)
				os.Remove(outputpath)
				os.Rename(outputmkimge , outputpath);

			}
			if imagesAttrs.ImagesArray[i].AttrVal == "0x00001001" {
				fmt.Println( "DTB Image" )
				outputmkimge := outputpath + ".mkimg"
				tmppath := outputpath + ".tmp"
				opath := "of=" + outputpath + ".tmp"
				iptah := "if=" + outputpath
				cmd = exec.Command("/bin/dd", iptah, opath, "bs=16", "conv=sync")
				cmd.Stderr = &stderr
				err = cmd.Run()
				if err != nil {
					return fmt.Errorf("AddHeader : Padding dtb image failed: %w %s", err, stderr.String())
				}
				cmd = exec.Command(path,"-A","x86","-O","linux","-C","none","-T",
				"flat_dt","-f","auto","-n","Flattened Device Tree",
				"-d",tmppath,outputmkimge)
				cmd.Stderr = &stderr
				err = cmd.Run()
				if err != nil {
					return fmt.Errorf("AddHeader : adding header to dtb image failed: %w %s",
					                  err, stderr.String())
				}
				//os.Remove(tmppath)
				os.Remove(outputpath)
				os.Rename(outputmkimge , outputpath);
			}
		}
	}
	return nil;
}

func splitHdrFromImage ( inputEncImgPath string, origImgPath string ) uint32 {
	fmt.Println("splitHdrFromImage")
	var hdrSize uint32
	var fullImgSize uint32
	var ImgSize uint32
	file, err := os.Open(inputEncImgPath)
	if err != nil {
		fmt.Println("opening encrypted image failed")
		return 0;
	}
	defer file.Close()
	fistat, err := file.Stat()
	if err != nil {
		return 0
	}
	fullImgSize = (uint32)(fistat.Size())
	fmt.Println("Full Image size:", fullImgSize)

	hdrSize = 128  //currently finding it fixed 128 bytes.if it will vary
			//then size calculation will be required
	fmt.Println("Attribute Header size:",hdrSize)
	inputimg := "if=" + inputEncImgPath
	cmd := exec.Command("dd", inputimg , "of=enc_hdr.bin" ,"bs=128", "count=1")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println("splitHdrFromImage : creating image header failed" + stderr.String())
		return 0
	}


	cmd = exec.Command("dd", inputimg , "of=enc_image.bin" ,"bs=1", "skip=128")
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println("splitHdrFromImage : creating image failed" + stderr.String())
		return 0
	}
	ImgSize = fullImgSize - hdrSize
	fmt.Println("Only image size:",ImgSize)
	return ImgSize
}

func CreateSBImage (inputEncImgPath string, origImgPath string, outputpath string,
                    sbImg types.SBImage, attrFile string, recovery int) error {

	sbImg.Type = "BLw"
	sbImg.Version = 2
	sbImg.PubKeyType = 2
	var stderr bytes.Buffer

	encImageSize := splitHdrFromImage(inputEncImgPath, origImgPath)
	if (encImageSize == 0) {
		return fmt.Errorf("CreateSBImage : splitting encrypted image and attr header failed")
	}
	file, err := os.Create(outputpath)
	defer file.Close()
	if err != nil {
		return fmt.Errorf("CreateSBImage : Creating output file failed : %w", err)
	}
	file.WriteString(sbImg.Type)
	type strtmp struct {
		ver   uint8
		pbKeyType uint32
	}
	pbKeybyte := make([]byte, 4)
	binary.BigEndian.PutUint32(pbKeybyte, uint32(sbImg.PubKeyType))
	for index := 0; index < len(pbKeybyte); index++ {
		pbKeybyte[0] = 1  //mappedSize 1 for ECC384
	}
	pbType := uint32(pbKeybyte[0])<<24 | uint32(pbKeybyte[1])<<16 | uint32(pbKeybyte[2]) <<8 | uint32(pbKeybyte[3])
	var s = strtmp{ sbImg.Version , pbType }
	err = binary.Write(file, binary.BigEndian, s)
	if err != nil {
		return fmt.Errorf("CreateSBImage : writing version and public key type failed : %w", err)
	}

	args := "cat " + sbImg.Signature + " >> " + outputpath
	fmt.Println(args)
	cmd := exec.Command("/bin/sh","-c", args)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("CreateSBImage : appening signature : %w %s", err, stderr.String())
	}
	//adding pblic key 96 bytes null
	const pbkey = "96"
	args = "dd if=/dev/zero " + "bs=1 " + "count=" + pbkey + " " + ">> " + outputpath
	fmt.Println(args)
	cmd = exec.Command("/bin/sh","-c", args)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("CreateSBImage : appending 96 null bytes for public key failed : %w %s", err, stderr.String())
	}
	//adding wrapped key and iv
	args = "cat " +  sbImg.EncryptionKey + " " + sbImg.IV + " >> " + outputpath
	fmt.Println(args)
	cmd = exec.Command("/bin/sh","-c", args)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("CreateSBImage : appending wrapped key and iv failed : %w %s", err, stderr.String())
	}

	file, err = os.OpenFile(outputpath,
		os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("CreateSBImage : opening file to write image size failed: %w", err)
	}
	defer file.Close()


	var imgSize uint32 = encImageSize
	err = binary.Write(file, binary.BigEndian, imgSize)
	if err != nil {
		return fmt.Errorf("CreateSBImage : writing image size failed : %w", err)
	}
	args = "cat " + "enc_hdr.bin" + " >> " + outputpath
	cmd = exec.Command("/bin/sh","-c", args)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("CreateSBImage : appending attribute header file failed : %w %s", err, stderr.String())
	}
	file, err = os.OpenFile(outputpath,
		os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("CreateSBImage : opening file to write image size failed: %w", err)
	}
	defer file.Close()
	var certCount uint32 = sbImg.CertCount
	err = binary.Write(file, binary.BigEndian, certCount)
	if err != nil {
		return fmt.Errorf("CreateSBImage : writing certificate count failed : %w", err)
	}
	args = "cat " + sbImg.Certificate + " >> " + outputpath
	fmt.Println(args)
	cmd = exec.Command("/bin/sh","-c", args)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("CreateSBImage : appending certificate failed : %w %s", err, stderr.String())
	}
	args = "cat " + "enc_image.bin" + " >> " + outputpath
	cmd = exec.Command("/bin/sh","-c", args)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("CreateSBImage : appending encrypted image  failed : %w %s", err, stderr.String())
	}

	err = AddHeader(attrFile, outputpath, origImgPath, recovery)
	if err != nil {
		return fmt.Errorf("CreateSBImage > %w", err)
	}

	return nil
}
