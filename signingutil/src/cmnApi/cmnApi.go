/*
 * Copyright (c) 2020-2022, MaxLinear, Inc.
 * 
 * APIs to read and add mkimage header,create fullimage
 *
 */
package cmnApi

import (
	"encoding/json"
	"fmt"
	"strings"
	"io/ioutil"
	"os"
	"os/exec"
	"bytes"
	"strconv"
	"errors"
	"path/filepath"
	"signingutil/types"
	"signingutil/sshSigning"
)

func ReadmkimageHdr(ImagePath string, imghdr *types.ImageHdrInfo) error {
	fmt.Println("Image header Information for image: " + ImagePath)
        exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("ReadmkimageHdr : Getting working directory failed: %w", err)
	}
        curdir := filepath.Dir(exe)
	path := curdir + "/mkimage"
	out, err := exec.Command(path, "-l", ImagePath).Output()
	if err != nil {
		return fmt.Errorf("ReadmkimageHdr : Getting mkimage header failed: %w", err)
	}
	if ( len(string(out)) <= 0 ) {
		err := errors.New("ReadmkimageHdr : mkimage header is empty")
		return err;
	}
        data := strings.Split(string(out), "\n")
        for i := range data {
		if strings.Contains(data[i], "FIT description") {
			imghdr.ImageType = "flat_dt"
			fmt.Println("It is a dtb image")
			return nil
		}
		if strings.Contains(data[i], "Image Name:") {
			ImageName := strings.TrimLeft(data[i],"Image Name:")
			imghdr.ImageName = strings.TrimSpace(ImageName)
			fmt.Printf("Image Name is [%s]\n", imghdr.ImageName)
		}
		if strings.Contains(data[i], "Type:") {
			ImageType := strings.TrimPrefix(data[i],"Type:")
			if strings.Contains(ImageType, "Filesystem") {
				imghdr.ImageType = "filesystem"
			} else if strings.Contains(ImageType, "Flat Device Tree") {
				imghdr.ImageType = "flat_dt"
			} else if strings.Contains(ImageType, "Kernel") {
				imghdr.ImageType = "kernel"
			} else if strings.Contains(ImageType, "uboot") ||
				strings.Contains(ImageType, "Boot") ||
				strings.Contains(ImageType, "FPGA") {
				imghdr.ImageType = "uboot"
			} else if strings.Contains(ImageType, "Firmware") {
				imghdr.ImageType = "firmware"
			}
			fmt.Printf("Image Type is [%s]\n", imghdr.ImageType)
			if strings.Contains(ImageType, "uncompressed") {
				imghdr.Compression = "none"
			} else if strings.Contains(ImageType, "compressed") {
				Compression := strings.ReplaceAll(ImageType, " compressed)", "")
				ss := strings.Split( Compression , "(" )
				imghdr.Compression = ss[len(ss)-1]
			}
			fmt.Printf("Image compression [%s]\n", imghdr.Compression)
		}
		if strings.Contains(data[i], "Load Address:") {
			Ldaddr := strings.TrimLeft(data[i],"Load Address:")
			strings.TrimSpace(Ldaddr)
			imghdr.LoadAddr = Ldaddr
			fmt.Printf("Load Address is [%s]\n", imghdr.LoadAddr)
		}
		if strings.Contains(data[i], "Entry Point:") {
			EntryPoint := strings.TrimLeft(data[i],"Entry Point:")
			strings.TrimSpace(EntryPoint)
			imghdr.Entrypoint = EntryPoint
			fmt.Printf("Entrypoint is [%s]\n", imghdr.Entrypoint)
		}
		if strings.Contains(data[i], "Compression:") {
			Compression := strings.TrimLeft(data[i]," Compression:")
			if strings.Contains(Compression, "uncompressed") {
                                imghdr.Compression = "none"
			}else{
				imghdr.Compression = strings.TrimSpace(Compression)
			}
			fmt.Printf("Image compression [%s]\n", imghdr.Compression)
		}
		if strings.Contains(data[i], "Image 0 ") {
			j := i + 1
			ImageName := strings.TrimLeft(data[j]," Description:")
			imghdr.ImageName = strings.TrimSpace(ImageName)
			fmt.Printf("Image Name is [%s]\n", imghdr.ImageName)
		}
	}
	return nil
}

func ReadServerConfig(servers *types.Servers) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("ReadServerConfig : getting executable failed: %w", err)
	}
	exePath := filepath.Dir(exe)
	//Reading server configuration
	serverjson := exePath + "/server.json"
	jFile, err := os.Open(serverjson)
	if err != nil {
		return fmt.Errorf("ReadServerConfig : Reading server configuration failed: %w", err)
	}
	defer jFile.Close()

	// read our opened xmlFile as a byte array.
        serversArray, _ := ioutil.ReadAll(jFile)

	// we unmarshal Array which contains our jsonFile's content into 'servers' which we defined above
	json.Unmarshal(serversArray, &servers)
	return nil
}
func Reconstruct_recovery( ImagePath string ) error {
	signedimgpath := ImagePath + ".signed"
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
func Reconstruct_rbe( imghdr types.ImageHdrInfo, ImagePath string ) error {
	signedimgpath := ImagePath + ".signed"
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
	cmd := exec.Command( swapbinpath ,signedimgpath,"rbe_temp.bin")
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

	cmd = exec.Command( swapbinpath ,"rbe_temp1.bin",outfile_tail)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Reconstruct_rbe : swap_bin.pl failed while creating rbe tail: %w %s",
			err, stderr.String())
	}
	/*create head*/
	opath := "of=" + outfile_head
	iptah := "if=" + strings.TrimSuffix(ImagePath, ".new")
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
	path := curdir + "/mkimage"
	cmd = exec.Command(path,"-A","x86","-C",imghdr.Compression,"-T",
			imghdr.ImageType,"-a",imghdr.LoadAddr,"-e",imghdr.Entrypoint,"-n",imghdr.ImageName,
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

func AddHeader ( imghdr types.ImageHdrInfo, ImagePath string ) error {
	fmt.Println("Adding image header...")
	var stderr bytes.Buffer

	signedimgpath := ImagePath + ".signed"
	ImageName := imghdr.ImageName

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("AddHeader : Getting working directory failed: %w", err)
	}
	curdir := filepath.Dir(exe)
	path := curdir + "/mkimage"
	uboot_mkimage := curdir + "/mkimage"
	pad2align := curdir + "/pad2align.sh"
	var cmd *exec.Cmd
	if imghdr.ImageType == "filesystem" {
		outputfile := signedimgpath + ".rtfsmkimg"
		tmppath := signedimgpath + ".tmp"
		opath := "of=" + signedimgpath + ".tmp"
		iptah := "if=" + signedimgpath
		cmd = exec.Command("/bin/dd", iptah, opath, "bs=16", "conv=sync")
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("AddHeader : Padding rootfs image failed: %w %s", err, stderr.String())
		}
		cmd = exec.Command(path,"-A","x86","-O","linux","-C",imghdr.Compression,"-T",
			imghdr.ImageType,"-a",imghdr.LoadAddr,"-e",imghdr.Entrypoint,"-n",ImageName,
			"-d",tmppath,outputfile)
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("AddHeader : adding header to rootfs image failed: %w %s",
				err, stderr.String())
		}
		os.Remove(tmppath)
		os.Remove(signedimgpath)

	}else if imghdr.ImageType == "flat_dt" {
		outputfile := signedimgpath + ".dtbmkimg"
		tmppath := signedimgpath + ".tmp"
		opath := "of=" + signedimgpath + ".tmp"
		iptah := "if=" + signedimgpath
		cmd = exec.Command("/bin/dd", iptah, opath, "bs=16", "conv=sync")
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("AddHeader : Padding dtb image failed: %w %s", err, stderr.String())
		}
	/*	cmd = exec.Command(path,"-A","x86","-O","linux","-C",imghdr.Compression,"-T",
			imghdr.ImageType,"-f","auto","-n",ImageName,"-d",tmppath,outputfile)*/

		cmd = exec.Command(path,"-A","x86","-O","linux","-C","none","-T",
                        "flat_dt","-f","auto","-n","'Flattened Device Tree'","-d",tmppath,outputfile)

		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("AddHeader : adding header to dtb image failed: %w %s",
				err, stderr.String())
		}
		os.Remove(tmppath)
		os.Remove(signedimgpath)
	}else if imghdr.ImageType == "kernel" {
		outputfile := signedimgpath + ".knlmkimg"


		tmppath := signedimgpath + ".tmp"
		opath := "of=" + signedimgpath + ".tmp"
		iptah := "if=" + signedimgpath
		cmd = exec.Command("/bin/dd", iptah, opath, "bs=16", "conv=sync")
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("AddHeader : Padding rootfs image failed: %w %s", err, stderr.String())
		}

		cmd = exec.Command(path,"-A","x86_64","-O","linux","-C",imghdr.Compression,"-T",
			imghdr.ImageType,"-a",imghdr.LoadAddr,"-e",imghdr.Entrypoint,"-n",ImageName,
			"-d",tmppath,outputfile)
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("AddHeader : adding header to dtb kernel image failed: %w %s",
				err, stderr.String())
		}
		os.Remove(tmppath)
		os.Remove(signedimgpath)
	} else if imghdr.ImageType == "uboot" {
		tmppath := signedimgpath + ".tmp"
		cmd = exec.Command( pad2align ,"-n","16",signedimgpath)
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("AddHeader : pad2align.sh failed to pad uboot image hdr: %w %s",
				err, stderr.String())
		}
		cmd = exec.Command(uboot_mkimage,"-A","x86","-C",imghdr.Compression,"-T",
			imghdr.ImageType,"-a",imghdr.LoadAddr,"-e",imghdr.Entrypoint,"-n",ImageName,
			"-d",signedimgpath,tmppath)
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("AddHeader : adding header to uboot image failed: %w %s",
				err, stderr.String())
		}
		os.Rename(tmppath,signedimgpath)
		os.Remove(tmppath)
	} else if imghdr.ImageType == "firmware" {
		err = Reconstruct_rbe( imghdr, ImagePath )
		if err != nil {
			return err;
		}
	} else if imghdr.ImageType == "recovery" {
		err = Reconstruct_recovery( ImagePath )
		if err != nil {
			return err;
		}
	} else if imghdr.ImageType == "tepfirmware" {
		outputfile := signedimgpath + ".mkimg"
		tmppath := signedimgpath + ".tmp"
		opath := "of=" + signedimgpath + ".tmp"
		iptah := "if=" + signedimgpath
		cmd = exec.Command("/bin/dd", iptah, opath, "bs=16", "conv=sync")
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("AddHeader : Padding tep firmware image failed: %w %s", err, stderr.String())
		}
		cmd = exec.Command(path,"-A","x86","-C",imghdr.Compression,"-T",
			"firmware","-n","TEP firmware",
			"-d",tmppath,outputfile)
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("AddHeader : adding header to tep firmware image failed: %w %s",
				err, stderr.String())
		}
		os.Remove(tmppath)
		os.Remove(signedimgpath)
	}
	return nil
}

func Sign(ImagePath string ,Username string, Passwrd string) error {
	var cmd *exec.Cmd
	var stderr bytes.Buffer
	var imghdr types.ImageHdrInfo
	var err error
	ImagePathNew := ImagePath + ".new"
	servercfgtype := "sbimage"

	//Reading server configuration
	var servers types.Servers
	err = ReadServerConfig(&servers)
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = ReadmkimageHdr(ImagePath, &imghdr)
	if err != nil {
		fmt.Println(err)
		if strings.Contains(ImagePath, "u-boot-mem") {
			imghdr.ImageType = "recovery"
		} else {
			fmt.Println(err)
			return err
		}
	}
	if strings.Contains(ImagePath, "u-boot-mem") {
		imghdr.ImageType = "recovery"
	}
	if strings.Contains(ImagePath,".dtb") {
		fmt.Println("Its dtb image")
		imghdr.ImageType = "flat_dt"
	}
	if imghdr.ImageType == "uboot" {
		fmt.Println("image type:" + imghdr.ImageType)
		tmppath := ImagePath + ".trim"
		opath := "of=" + ImagePath + ".trim"
		iptah := "if=" + ImagePath
		cmd = exec.Command("/bin/dd", iptah, opath, "bs=1", "skip=64")
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			fmt.Println("Sign : Removing header from uboot failed" + stderr.String())
			return err
		}
		os.Rename( tmppath, ImagePathNew)
	} else if imghdr.ImageType == "firmware" {
		if imghdr.ImageName == "TepFW" {
			imghdr.ImageType = "tepfirmware"
			fmt.Println("[TEP] image type:" + imghdr.ImageType)
			tmppath := ImagePath + ".trim"
			opath := "of=" + ImagePath + ".trim"
			iptah := "if=" + ImagePath
			cmd = exec.Command("/bin/dd", iptah, opath, "bs=1", "skip=64")
			cmd.Stderr = &stderr
			err = cmd.Run()
			if err != nil {
				fmt.Println("Sign : Removing header failed" + stderr.String())
				return err
			}
			os.Rename( tmppath, ImagePathNew)
		} else if imghdr.ImageName == "RBE" {
			fmt.Println("image type:" + imghdr.ImageType)
			args := "/usr/bin/xxd -p -l4 -s 0x844 " + ImagePath + " | " + " /usr/bin/xxd -r -p > fw_temp.bin"
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
			tmppath := ImagePath + ".trim"
			opath := "of=" + ImagePath + ".trim"
			iptah := "if=" + ImagePath
			count := "count=" + strconv.FormatInt(rbesize, 10)
			cmd = exec.Command("/bin/dd", iptah, opath, "bs=1", "skip=2124", count)
			cmd.Stderr = &stderr
			err = cmd.Run()
			if err != nil {
				fmt.Println("Sign : Removing header from rbe image failed" + stderr.String())
				return err
			}
			os.Rename( tmppath, ImagePathNew)
			os.Remove("fw_temp.bin")
			os.Remove("fw_temp1.bin")
		}
	} else if imghdr.ImageType == "kernel" ||
		imghdr.ImageType == "filesystem" {
		fmt.Println("image type:" + imghdr.ImageType)
		tmppath := ImagePath + ".trim"
		opath := "of=" + ImagePath + ".trim"
		iptah := "if=" + ImagePath
		cmd = exec.Command("/bin/dd", iptah, opath, "bs=1", "skip=64")
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			fmt.Println("Sign : Removing header failed" + stderr.String())
			return err
		}
		os.Rename( tmppath, ImagePathNew)

	} else if imghdr.ImageType == "flat_dt" ||
		imghdr.ImageType == "recovery" {
		fmt.Println("image type:" + imghdr.ImageType)
                cmd = exec.Command("/bin/cp", ImagePath, ImagePathNew)
                cmd.Stderr = &stderr
                err = cmd.Run()
                if err != nil {
                        fmt.Println("copy file failed" + stderr.String())
                        return err
                }
	}
	for i := 0; i < len(servers.Servers); i++ {
		if ((servers.Servers[i].Type == servercfgtype) &&
			(servers.Servers[i].UserName == Username)) {
			err = sshSigning.SignImage( servers.Servers[i], ImagePathNew, imghdr, Passwrd )
			if err != nil {
				return err
			}
		}
	}
	//Adding mkimage header
	err = AddHeader( imghdr, ImagePathNew )
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("done")
	os.Remove(ImagePathNew)
	return nil
}
func ROTSign (ImagePath string ,Username string, Passwrd string) error {
	var servers types.Servers
	var imghdr types.ImageHdrInfo
	imghdr.ImageType = "rot"
	err := ReadServerConfig(&servers)
	if err != nil {
		fmt.Println(err)
		return err
	}
	for index := 0; index < len(servers.Servers); index++ {
		if ((servers.Servers[index].Type == "rot") &&
			(servers.Servers[index].UserName == Username)) {
			err = sshSigning.SignImage( servers.Servers[index], ImagePath, imghdr, Passwrd )
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func GenCustImgHdr ( ImagePath string ) error {
	const (
		custHdrOffset = "0x80000"
		hdrAlignbyte = "396"
		crcAlignbyte = "2048"
        )
	fmt.Println("GenCustImgHdr")
	custImageHdr := ImagePath + "cust-img-header.bin"
	fmt.Println(custImageHdr)
	//Generating Customer Test Image Header @ 0x80000 .......	
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("GenCustImgHdr : Getting working directory failed: %w", err)
	}
	curdir := filepath.Dir(exe)
	/*create tail*/
	swapbinpath := curdir + "/swap_bin.pl"
	hostcrc32 := curdir + "/hostcrc32"
	mkheadpath := curdir + "/mk_intel_header.pl"
	pad2align := curdir + "/pad2align.sh"

	cmd := exec.Command( "/usr/bin/touch" ,"dummy")
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("GenCustImgHdr : creating dummy file failed: %w", err)
	}
	cmd = exec.Command( mkheadpath ,"-ddr","dummy","--offset",custHdrOffset,
			"--out","cust-ibrh.tmp")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("GenCustImgHdr : mk_intel_header.pl failed to create customer image hdr: %w %s",
			err, stderr.String())
	}
	cmd = exec.Command( pad2align ,"-n",hdrAlignbyte,"cust-ibrh.tmp" )
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("GenCustImgHdr : pad2align.sh failed to pad customer image hdr: %w %s",
			err, stderr.String())
	}
	cmd = exec.Command( swapbinpath ,"cust-ibrh.tmp","cust-ibrh-swap.tmp" )
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("GenCustImgHdr : swap_bin.pl failed to create customer image hdr: %w %s",
			err, stderr.String())
	}
	cmd = exec.Command( hostcrc32 ,"cust-ibrh-swap.tmp","cust-ibrh-swap-crc.tmp" )
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("GenCustImgHdr : hostcrc32 failed to encrypt customer image hdr: %w %s",
			err, stderr.String())
	}
	cmd = exec.Command( pad2align ,"-n",crcAlignbyte,"cust-ibrh-swap-crc.tmp" )
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("GenCustImgHdr : pad2align.sh failed to pad customer image hdr: %w %s",
			err, stderr.String())
	}
	fmt.Println("GenCustImgHdr" + custImageHdr);
	err = os.Rename( "cust-ibrh-swap-crc.tmp"  , custImageHdr )
	if err != nil {
		fmt.Println("rename failed");
		fmt.Println(err);
	}
	os.Remove( "cust-ibrh.tmp")
	os.Remove( "cust-ibrh-swap.tmp")
	os.Remove( "cust-ibrh-swap-crc.tmp")
	return nil
}

func GenManuHdr ( ImagePath string ) error {
	const (
		manuHdrOffset = "0x100000"
		hdrAlignbyte = "396"
		crcAlignbyte = "2048"
        )
	manu_image_hdr := ImagePath + "manuboot-img-header.bin"
	//Generating Manuboot Image Header @ 0x100000 .......	
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("GenManuHdr : Getting working directory failed: %w", err)
	}
	curdir := filepath.Dir(exe)
	/*create tail*/
	swapbinpath := curdir + "/swap_bin.pl"
	hostcrc32 := curdir + "/hostcrc32"
	mkheadpath := curdir + "/mk_intel_header.pl"
	pad2align := curdir + "/pad2align.sh"

	cmd := exec.Command( "/usr/bin/touch" ,"dummy")
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("GenManuHdr : creating dummy file failed: %w", err)
	}
	cmd = exec.Command( mkheadpath ,"-ddr","dummy","--offset",manuHdrOffset,
			"--out","manuboot-ibrh.tmp")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("GenManuHdr : mk_intel_header.pl failed to create manuboot image hdr: %w %s",
			err, stderr.String())
	}
	cmd = exec.Command( pad2align ,"-n",hdrAlignbyte,"manuboot-ibrh.tmp" )
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("GenManuHdr : pad2align.sh failed to pad manuboot image hdr: %w %s",
			err, stderr.String())
	}
	cmd = exec.Command( swapbinpath ,"manuboot-ibrh.tmp","manuboot-ibrh-swap.tmp" )
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("GenManuHdr : swap_bin.pl failed to create manuboot swap hdr: %w %s",
			err, stderr.String())
	}
	cmd = exec.Command( hostcrc32 ,"manuboot-ibrh-swap.tmp","manuboot-ibrh-swap-crc.tmp" )
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("GenManuHdr : hostcrc32 failed to encrypt manuboot image hdr: %w %s",
			err, stderr.String())
	}
	cmd = exec.Command( pad2align ,"-n",crcAlignbyte,"manuboot-ibrh-swap-crc.tmp" )
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("GenManuHdr : pad2align.sh failed to pad manuboot image hdr: %w %s",
			err, stderr.String())
	}
	os.Rename( "manuboot-ibrh-swap-crc.tmp"  , manu_image_hdr )
	os.Remove( "manuboot-ibrh.tmp")
	os.Remove( "manuboot-ibrh-swap.tmp")
	os.Remove( "manuboot-ibrh-swap-crc.tmp")
	return nil
}
func FieldImgGen ( ImagePath string, FieldImgPath string ) error {

	fieldIPath := ImagePath + "field-img.post"
	fieldImgHdr := ImagePath + "field-img-header.bin"
	fieldImgTail := ImagePath + "field-img-tail.bin"
	opath := "of=" + ImagePath + "field-img.post"
	iptah := "if=" + FieldImgPath
	cmd := exec.Command("/bin/dd", iptah, opath, "bs=1", "skip=64")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("FieldImgGen : Removing mkimage header failed" + stderr.String())
		return err
	}
	cmd = exec.Command("/usr/bin/split", "-b", "2048" , fieldIPath , "split_field_img_")
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println("FieldImgGen : splitting field image to read header failed" + stderr.String())
		return err
	}
	os.Rename( "split_field_img_aa"  , fieldImgHdr )
	args := "cat " + "split_field_img_* " + " > " + fieldImgTail
        cmd = exec.Command("/bin/sh","-c", args)
        cmd.Stderr = &stderr
        err = cmd.Run()
        if err != nil {
		return fmt.Errorf("FieldImgGen : Field image tail creation failed: %w %s",
			err, stderr.String())
        }
	args = "rm" + " -rf " + "split_field_img_*"
	fmt.Println("Removing file" + args)
	cmd = exec.Command("/bin/sh","-c", args)
        cmd.Stderr = &stderr
        err = cmd.Run()
        if err != nil {
        }
	return nil
}
func CustImgGen ( ImagePath string, CustImgPath string ) error {
	custImgTail := ImagePath + "cust-img-tail.bin"
	cmd := exec.Command("/usr/bin/split", "-b", "2048" , CustImgPath , "split_cust_img_")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("CustImgGen : splitting cust image failed" + stderr.String())
		return err
	}
	os.Rename( "split_cust_img_aa"  , "cust-img-u-boot-ibrh.tmp" )
	args := "cat " + "split_cust_img_* " + " > " + custImgTail
        cmd = exec.Command("/bin/sh","-c", args)
        cmd.Stderr = &stderr
        err = cmd.Run()
        if err != nil {
		return fmt.Errorf("CustImgGen : Customer image tail creation failed: %w %s",
			err, stderr.String())
        }
	args = "rm" + " -rf " + "split_cust_img_*"
	fmt.Println("Removing file" + args)
	cmd = exec.Command("/bin/sh","-c", args)
        cmd.Stderr = &stderr
        err = cmd.Run()
        if err != nil {
        }
	os.Remove( "cust-img-u-boot-ibrh.tmp")
	return nil
}

func ManubootImgGen ( ImagePath string, ManuImgPath string ) error {

	const (
		manuboot_addr = "0xBFF20000"
		socmode = "1"
	)
	manuImgPathSigned := ManuImgPath + ".signed"
	fmt.Println("$$$$$$$ManubootImgGen:" + manuImgPathSigned)
	manuImgTail := ImagePath + "manuboot-img-tail.bin"
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("ManubootImgGen : Getting working directory failed: %w", err)
	}
	curdir := filepath.Dir(exe)
	/*create tail*/
	swapbinpath := curdir + "/swap_bin.pl"
	mktailpath := curdir + "/mk_intel_tail.pl"
	cmd := exec.Command( swapbinpath ,manuImgPathSigned,"manuboot-swap.tmp")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("ManubootImgGen : swap_bin.pl failed while creating manuboot image: %w %s",
			err, stderr.String())
	}
	cmd = exec.Command( mktailpath ,"-i","manuboot-swap.tmp","-o",
		"manuboot-tail.tmp", "-j",manuboot_addr,"-m",socmode)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("ManubootImgGen : mk_intel_tail.pl failed while creating manuboot image: %w %s",
			err, stderr.String())
	}
	cmd = exec.Command( swapbinpath ,"manuboot-tail.tmp",manuImgTail)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("ManubootImgGen : swap_bin.pl failed while creating manuboot image: %w %s",
			err, stderr.String())
	}
	os.Remove( "manuboot-swap.tmp")
	os.Remove( "manuboot-tail.tmp")
	return nil
}
func ROTFullImgGen ( ImagePath string ) error {
	const (
		pad2aligncusthdr = "0x80000"
		pad2aligncustimg = "0x100000"
		fieldibrh = "396"
		crcalignbyte = "2048"
        )
	fieldImg := ImagePath + "field-img.post"
	fieldImgHdr := ImagePath + "field-img-header.bin"
	fieldImgTail := ImagePath + "field-img-tail.bin"
	rotfullimage := ImagePath + "rot-full-img-emmc.bin"
	custImageHdr := ImagePath + "cust-img-header.bin"
	custImageTail := ImagePath + "cust-img-tail.bin"
	manubootHdr := ImagePath + "manuboot-img-header.bin"
	manubootTail := ImagePath + "manuboot-img-tail.bin"

	os.Remove(rotfullimage)
	os.Rename(fieldImg,rotfullimage)

	//tmppath := signedimgpath + ".tmp"
	opath := "of=" + rotfullimage
	iptah := "if=" + custImageHdr
	cmd := exec.Command("/bin/dd", iptah, opath, "conv=notrunc")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("ROTFullImgGen : adding cust hdr to ROT image failed: %w %s", err, stderr.String())
	}
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("ROTFullImgGen : Getting working directory failed: %w", err)
	}
	curdir := filepath.Dir(exe)
	/*create tail*/
	pad2align := curdir + "/pad2align.sh"
	cmd = exec.Command( pad2align ,"-n",pad2aligncusthdr,rotfullimage )
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("ROTFullImgGen : pad2align.sh failed to pad ROT Fullimage after cust hdr: %w %s",
			err, stderr.String())
	}
	args := "cat " + custImageTail + " >> " + rotfullimage
	cmd = exec.Command("/bin/sh","-c", args)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("ROTFullImgGen : appending cust image to ROT full image failed : %w", err)
	}
	cmd = exec.Command( pad2align ,"-n",pad2aligncustimg,rotfullimage )
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("ROTFullImgGen : pad2align.sh failed to pad ROT Fullimage after cust image: %w %s",
			err, stderr.String())
	}
	// location of field ibrh @ 0x100000 - 0x800  = 0xFF800 (1046528) #
	opath = "of=" + rotfullimage
	iptah = "if=" + fieldImgHdr
	cmd = exec.Command("/bin/dd", iptah, opath, "conv=notrunc" , "bs=1","seek=1046528")
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("ROTFullImgGen : adding field image hdr to ROT image failed: %w %s", err, stderr.String())
	}
	//# location of manuboot ibrh @ 0x100000 - 0x1000  = 0xFF000 #
	opath = "of=" + rotfullimage
	iptah = "if=" + manubootHdr
	cmd = exec.Command("/bin/dd", iptah, opath, "conv=notrunc","bs=1","seek=1044480")
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("ROTFullImgGen : adding manuboot hdr to ROT image failed: %w %s", err, stderr.String())
	}
	args = "cat " + manubootTail + " >> " + rotfullimage
	cmd = exec.Command("/bin/sh","-c", args)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("ROTFullImgGen : appending manuboot image to ROT full image failed : %w", err)
	}
	os.Remove(fieldImgHdr)
	os.Remove(fieldImgTail)
	os.Remove(custImageHdr)
	os.Remove(custImageTail)
	os.Remove(manubootHdr)
	os.Remove(manubootTail)
	return nil

}
func CreateROTFullimage(Manuboot string, CustImage string,
			FieldImage string) error {
	var cmd *exec.Cmd
	var stderr bytes.Buffer
	fmt.Println( "creating ROT fullimage :")

	workdir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	ImagePath := workdir + "/" + "rot_output" + "/"
	os.RemoveAll(ImagePath)
	cmd = exec.Command("/bin/mkdir", ImagePath)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println("Sign : Removing header failed" + stderr.String())
		return err
	}
	err = GenCustImgHdr(ImagePath)
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = GenManuHdr(ImagePath)
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = FieldImgGen(ImagePath , FieldImage)
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = CustImgGen(ImagePath , CustImage)
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = ManubootImgGen(ImagePath , Manuboot)
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = ROTFullImgGen(ImagePath)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
func FullimageCreation(fullimgpath string, rootfsimgpath string,
			dtbimgpath string, krnlimgpath string) error {

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("CreateFullimage : Getting working directory failed: %w", err)
	}
	curdir := filepath.Dir(exe)
	path := curdir + "/mkimage"
	infilepath := curdir + "/fullimage.img.tmp"

	args := "cat " + krnlimgpath + " " + rootfsimgpath + " " + dtbimgpath + " > " + infilepath
        cmd := exec.Command("/bin/sh","-c", args)
	fmt.Println(args)
	var stderr bytes.Buffer
        cmd.Stderr = &stderr
        err = cmd.Run()
        if err != nil {
		return fmt.Errorf("CreateFullimage : Creating fullimage failed: %w %s",
			err, stderr.String())
        }
	cmd = exec.Command(path,"-A","x86","-O","linux","-C","none","-T","multi","-a",
		"0x00","-e","0x00","-n","OpenWrt fullimage","-d",infilepath,fullimgpath)
        cmd.Stderr = &stderr
        err = cmd.Run()
        if err != nil {
		return fmt.Errorf("CreateFullimage : Adding header to fullimage failed: %w %s",
			err, stderr.String())
        }
	os.Remove(infilepath)
	return nil;
}
func CreateFullimage(ImagePath string , FullImgCfgName string) error {
	var krnlimgpath string
	var rootfsimgpath string
	var dtbimgpath string
	var dtbcount int
	var fullimgpath string
	fmt.Println( "creating fullimage :")
	files, err := ioutil.ReadDir( ImagePath )
	if err != nil {
		return fmt.Errorf("CreateFullimage : Getting working directory failed: %w", err)
	}
	//read each file an create fullimage
	for _, file := range files {
		var imgfullpath string
		//fmt.Println(file.Name(), file.IsDir())
		strLastChar := ImagePath
		if strLastChar[len(strLastChar)-1] == '/' {
			imgfullpath = ImagePath + file.Name()
		} else {
			imgfullpath = ImagePath + "/" + file.Name()
		}
		if strings.Contains(imgfullpath, "signed.rtfsmkimg") {
			rootfsimgpath =  imgfullpath
			rootfsimgpath = strings.ReplaceAll(rootfsimgpath, ".new.signed.rtfsmkimg", "")
			os.Rename( imgfullpath , rootfsimgpath)
		}
		if strings.Contains(imgfullpath, ".dtbmkimg") {
			dtbcount = dtbcount + 1
			dtbimgpath =  imgfullpath
			dtbimgpath = strings.ReplaceAll(dtbimgpath , ".new.signed.dtbmkimg", "")
			os.Rename( imgfullpath , dtbimgpath)
		}
		if strings.Contains(imgfullpath, "signed.knlmkimg") {
			krnlimgpath =  imgfullpath
			krnlimgpath = strings.ReplaceAll(krnlimgpath , ".new.signed.knlmkimg", "")
			os.Rename( imgfullpath , krnlimgpath)
		}
	}
	if ( len ( rootfsimgpath ) <= 0 ) {
		err := errors.New("CreateFullimage : signed rootfs image not found")
		fmt.Println(err)
		return err;
	}
	if ( len ( dtbimgpath ) <= 0 ) {
		err := errors.New("CreateFullimage : signed dtb image not found")
		fmt.Println(err)
		return err;
	}
	if ( len ( krnlimgpath ) <= 0 ) {
		err := errors.New("CreateFullimage : signed kernel image not found")
		fmt.Println(err)
		return err;
	}
	if ( dtbcount > 1 ){
		var imagescfg types.ImagesCfg
		jFile, err := os.Open(FullImgCfgName)
		if err != nil {
			return fmt.Errorf("CreateFullimage : Configuration file specifyin dtb and fullimage name is missing: %w",err)
		}
		defer jFile.Close()
		imgCfgArray, _ := ioutil.ReadAll(jFile)
		json.Unmarshal(imgCfgArray, &imagescfg)
		for i := 0; i < len(imagescfg.FullImages); i++ {
			fullimgpath = ImagePath + "/" + imagescfg.FullImages[i].FullImgName
			dtbimgpath = ImagePath + "/" + imagescfg.FullImages[i].DtbName
			if _, err = os.Stat( dtbimgpath )
			err != nil {
				continue
			}
			err = FullimageCreation(fullimgpath,rootfsimgpath,dtbimgpath,
					krnlimgpath)
			if err != nil {
				return fmt.Errorf("CreateFullimage : Fullimage creation failed: %w", err)
			}
                }
        } else {
		fullimgpath = ImagePath + "/" + "fullimage.img"
		err = FullimageCreation(fullimgpath,rootfsimgpath,dtbimgpath,
					krnlimgpath)
		if err != nil {
			return fmt.Errorf("CreateFullimage : Fullimage creation failed: %w", err)
		}
	}
	return nil
}

func CleanupTempFiles ( tmpfilepath ...string ) {

	for _, fpath := range tmpfilepath {
		if _, err := os.Stat(fpath);
		err == nil {
			os.Remove(fpath)
		}
	}
	mountPath :="/home/ext4_kernel_rootfs"
	args := "sudo " + "rm " + "-rf " + mountPath
	fmt.Println(args)
	cmd := exec.Command("/bin/sh", "-c", args)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Run()
}

func CreateUpTotalImg(ubootImg string ,tepfwImg string ,dtbImg string , kernelImg string ,
	rootfsImg string ,outputImg string) error {

	const (
		ubootpartsize = "1048576"
		tepfwpartsize = "2097152"
		dtbpartsize = "524288"
		extrapadding = "1572864"
	)

	var stderr bytes.Buffer
	//uboot image
	ubootPaddedImg := ubootImg + ".pad"
	cmd := exec.Command("/bin/cp",ubootImg,ubootPaddedImg)
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("CreateUpTotalImg : copying uboot image to pad failed : %w", err)
	}
	ubootPad := "of=" + ubootPaddedImg
	ubootseek := "seek=" + ubootpartsize
	cmd = exec.Command("/bin/dd","if=/dev/null",ubootPad,"obs=1",ubootseek)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		CleanupTempFiles ( ubootPaddedImg )
		return fmt.Errorf("CreateUpTotalImg : padding uboot image to 1M failed : %w", err)
	}
	//tepfw image
	tepfwPaddedImg := tepfwImg + ".pad"
	cmd = exec.Command("/bin/cp", tepfwImg, tepfwPaddedImg)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		CleanupTempFiles ( ubootPaddedImg )
		return fmt.Errorf("CreateUpTotalImg : copying tepfw image to pad failed : %w", err)
	}
	//stripping mkimage header if it present
	exe, err := os.Executable()
	if err != nil {
		CleanupTempFiles ( ubootPaddedImg ,tepfwPaddedImg )
		return fmt.Errorf("CreateUpTotalImg : Getting working directory failed: %w", err)
	}
	curdir := filepath.Dir(exe)
	path := curdir + "/mkimage"
	out, err := exec.Command(path, "-l", tepfwPaddedImg).Output()
	if (( err == nil ) && ( len(string(out)) > 0 ) &&
		strings.Contains(string(out), "Type") ) {
		fmt.Println("CreateUpTotalImg : mkimage header is present in TEP Firmware , strpping it")
		tmppath := tepfwPaddedImg + ".trim"
		opath := "of=" + tepfwPaddedImg + ".trim"
		iptah := "if=" + tepfwPaddedImg
		cmd = exec.Command("/bin/dd", iptah, opath, "bs=1", "skip=64")
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			CleanupTempFiles ( ubootPaddedImg ,tepfwPaddedImg )
			fmt.Println("CreateUpTotalImg : Removing header from tetp firmware failed" + stderr.String())
			return err
		}
		os.Rename( tmppath, tepfwPaddedImg)
	}
	//padding tep firmware
	tepfwPad := "of=" + tepfwPaddedImg
	tefwseek := "seek=" + tepfwpartsize
	cmd = exec.Command("/bin/dd","if=/dev/null",tepfwPad,"obs=1",tefwseek)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		CleanupTempFiles ( ubootPaddedImg ,tepfwPaddedImg )
		return fmt.Errorf("CreateUpTotalImg : padding tepfirmware image to 2M failed : %w", err)
	}
	//dtb image
	dtbPaddedImg := dtbImg + ".pad"
	cmd = exec.Command("/bin/cp", dtbImg, dtbPaddedImg)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		CleanupTempFiles ( ubootPaddedImg ,tepfwPaddedImg )
		return fmt.Errorf("CreateUpTotalImg : copying dtb image to pad failed : %w", err)
	}
	dtbfwPad := "of=" + dtbPaddedImg
	dtbseek := "seek=" + dtbpartsize
	cmd = exec.Command("/bin/dd","if=/dev/null",dtbfwPad,"obs=1",dtbseek)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		CleanupTempFiles ( ubootPaddedImg ,tepfwPaddedImg ,dtbPaddedImg)
		return fmt.Errorf("CreateUpTotalImg : padding dtb image to 512KB failed : %w", err)
	}

	//Creating total image

	//adding gpt table
	gptHeaderpad := "gptheader" + ".pad"
	cmd = exec.Command("/bin/cp", "gptheader" , gptHeaderpad)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		CleanupTempFiles ( ubootPaddedImg ,tepfwPaddedImg ,dtbPaddedImg)
		return fmt.Errorf("CreateUpTotalImg : copying GPT header to pad failed : %w", err)
	}
	gptHeaderof := "of=" + gptHeaderpad
	cmd = exec.Command("/bin/dd","if=/dev/null", gptHeaderof,"obs=1",dtbseek)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		CleanupTempFiles ( ubootPaddedImg ,tepfwPaddedImg ,dtbPaddedImg , gptHeaderpad)
		return fmt.Errorf("CreateUpTotalImg : padding GPT header to 512KB failed : %w", err)
	}

	args := "cat " + gptHeaderpad + " " + ubootPaddedImg + " " + ubootPaddedImg + " "  + 
	tepfwPaddedImg + " " + tepfwPaddedImg + " " + dtbPaddedImg + " " + dtbPaddedImg + " > " + outputImg
	fmt.Println(args)
	cmd = exec.Command("/bin/sh","-c", args)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		CleanupTempFiles ( ubootPaddedImg ,tepfwPaddedImg ,dtbPaddedImg , gptHeaderpad)
		return fmt.Errorf("CreateUpTotalImg : Creating total image failed : %w", err)
	}
	//args = "/bin/dd" + " if=/dev/zero " + "bs=1" + " count=524288 >> " + outputImg  
	// padding 512 + 512 + 512 kb (uboot_env_a + uboot_env_b + extrapadding before extended_uboot_a)
	args = "/bin/dd" + " if=/dev/zero " + "bs=1" + " count=" + extrapadding + " >> " + outputImg
	fmt.Println(args)
	cmd = exec.Command("/bin/sh","-c", args)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		CleanupTempFiles ( ubootPaddedImg ,tepfwPaddedImg ,dtbPaddedImg , gptHeaderpad)
		return fmt.Errorf("CreateUpTotalImg : padding 512 KB for uboot config failed : %w", err)
	}

	/*Handling extended partition*/

	//stripping mkimage header if it is present
	rootfsImgtmp := rootfsImg + ".pad"
	cmd = exec.Command("/bin/cp" ,rootfsImg,rootfsImgtmp)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		CleanupTempFiles ( ubootPaddedImg ,tepfwPaddedImg ,dtbPaddedImg , gptHeaderpad)
		return fmt.Errorf("CreateUpTotalImg : copying rootfs image failed : %w", err)
	}
	out, err = exec.Command(path, "-l", rootfsImgtmp).Output()
	if (( err == nil ) && ( len(string(out)) > 0 ) &&
		strings.Contains(string(out), "Type") ) {
		fmt.Println("CreateUpTotalImg : mkimage header is present in rootfs , strpping it")
		tmppath := rootfsImgtmp + ".trim"
		opath := "of=" + rootfsImgtmp + ".trim"
		iptah := "if=" + rootfsImgtmp
		cmd = exec.Command("/bin/dd", iptah, opath, "bs=1", "skip=64")
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			CleanupTempFiles ( ubootPaddedImg ,tepfwPaddedImg ,dtbPaddedImg , gptHeaderpad,
					rootfsImgtmp )
			fmt.Println("CreateUpTotalImg : Removing header from rootfs failed" + stderr.String())
			return err
		}
		os.Rename( tmppath, rootfsImgtmp)
	}

	//mounting ext4.fs and handling kernel rootfs there
	mountPath :="/home/ext4_kernel_rootfs"
	args = "sudo " + "mkdir " + mountPath
	fmt.Println(args)
	cmd = exec.Command("/bin/sh","-c", args)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		CleanupTempFiles ( ubootPaddedImg ,tepfwPaddedImg ,dtbPaddedImg , gptHeaderpad,
				rootfsImgtmp )
		return fmt.Errorf("CreateUpTotalImg : creating /home/qateam/ext4_kernel_rootfs : %w %s", err,stderr.String())
	}
	args = "sudo " + "mount -t ext4 " + " ext4.fs "  + mountPath
	fmt.Println(args)
	cmd = exec.Command("/bin/sh","-c", args)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		CleanupTempFiles ( ubootPaddedImg ,tepfwPaddedImg ,dtbPaddedImg , gptHeaderpad,
				rootfsImgtmp )
		return fmt.Errorf("CreateUpTotalImg : mount failed : %w %s", err,stderr.String())
	}

	args = "sudo " + "/usr/bin/cp " + kernelImg + " " + mountPath
	fmt.Println(args)
	cmd = exec.Command("/bin/sh","-c", args)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		CleanupTempFiles ( ubootPaddedImg ,tepfwPaddedImg ,dtbPaddedImg , gptHeaderpad,
				rootfsImgtmp )
		return fmt.Errorf("CreateUpTotalImg : copying kernel to mounted ext4 filesystem : %w %s", err,stderr.String())
	}

	lastInd := strings.LastIndex( rootfsImg, "/" )
	imgnm := rootfsImg[lastInd+1:]
	args = "sudo " + "/usr/bin/cp " + rootfsImgtmp + " " + mountPath + "/" + imgnm
	fmt.Println(args)
	cmd = exec.Command("/bin/sh","-c", args)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		CleanupTempFiles ( ubootPaddedImg ,tepfwPaddedImg ,dtbPaddedImg , gptHeaderpad,
				rootfsImgtmp )
		return fmt.Errorf("CreateUpTotalImg : copying rootfs to mounted ext4 filesystem : %w %s", err,stderr.String())
	}

	args = "sudo " + "/usr/bin/umount " + "ext4.fs"
	cmd = exec.Command("/bin/sh","-c", args)
	fmt.Println(args)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		CleanupTempFiles ( ubootPaddedImg ,tepfwPaddedImg ,dtbPaddedImg , gptHeaderpad,
				rootfsImgtmp )
		return fmt.Errorf("CreateUpTotalImg : umount failed : %w %s", err,stderr.String())
	}

	//adding kernel and rootfs
	args = "cat " + "ext4.fs" + " " + "ext4.fs" + " >> " + outputImg
	fmt.Println(args)
	cmd = exec.Command("/bin/sh","-c", args)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		CleanupTempFiles ( ubootPaddedImg ,tepfwPaddedImg ,dtbPaddedImg , gptHeaderpad,
				rootfsImgtmp )
		return fmt.Errorf("CreateUpTotalImg : appending kernel rootfs in total image failed : %w %s", err,stderr.String())
	}
	CleanupTempFiles ( ubootPaddedImg ,tepfwPaddedImg ,dtbPaddedImg , gptHeaderpad,
			rootfsImgtmp )
	return nil
}

