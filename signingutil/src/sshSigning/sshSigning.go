/*
 * Copyright (c) 2020-2022, MaxLinear, Inc.
 * 
 * This file handles signing of the image
 *
 */

package sshSigning

import (
	"fmt"
	"os/exec"
	"encoding/json"
	"strings"
	"io/ioutil"
	"os"
	"path/filepath"
	"signingutil/types"
	"bytes"
)

func CopyImagetoServer(serverinfo types.Server,ImagePath string,
			Passwrd string ) error {
	fmt.Println("Copying image to server...")
	ServInfo := serverinfo.UserName + "@" + serverinfo.Host + ":" + serverinfo.SigntoolPath
	//fmt.Println("command " + ServInfo)
	cmd := exec.Command("sshpass", "-p", Passwrd, "scp", ImagePath, ServInfo)
	err := cmd.Run()
	if err != nil {
		cmd := exec.Command("sshpass", "-p", Passwrd, "scp", ImagePath, ServInfo)
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("CopyImagetoServer : copying image to server failed: %w", err)
		}
	}
	return nil
}

func CopyImageFromServer(serverinfo types.Server,ImagePath string,
			Passwrd string ) error {
	fmt.Println("Copying image from server...")
	var imgnm string
	var path string
	lastInd := strings.LastIndex( ImagePath, "/" )
	fmt.Println(lastInd)
	if lastInd >= 0 {
		imgnm = ImagePath[lastInd+1:]
		path = ImagePath[:lastInd] + "/"
	} else {
		imgnm = ImagePath
		path = "."
	}
	fmt.Println(imgnm)
	fmt.Println(path)

	ServInfo := serverinfo.UserName + "@" + serverinfo.Host + ":" + serverinfo.SigntoolPath + "/" + imgnm + ".signed"
	cmd := exec.Command("sshpass", "-p", Passwrd, "scp", ServInfo, path)
	err := cmd.Run()
	if err != nil {
		cmd := exec.Command("sshpass", "-p", Passwrd, "scp", ServInfo, path)
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("CopyImagetoServer : copying image from server failed: %w", err)
		}
	}
	return nil
}

func Imagesigning(ImageName string, imageattr types.ImageAttr , serverinfo types.Server,
			Passwrd string ) error {
	fmt.Println("Signing image..." + ImageName)
	var args string
	outfilename := ImageName + ".signed"
	srvinfo := serverinfo.UserName + "@" + serverinfo.Host
	privkey := serverinfo.PrivKey
	wrapkey := serverinfo.WrapKey
	cert := serverinfo.Cert
	signtoolpath := serverinfo.SigntoolPath + "/" + "signtool"
	if imageattr.Type == "rot" {
		featurecred := serverinfo.FeatureCred
		otpkeyblob := serverinfo.OtpKeyBlob
		fmt.Println("inside ROT sign")
		args = "sign" + " -type" + " BLw" + " -pubkeytype" + " otp" + " -prikey " + privkey +
			" -infile " + serverinfo.SigntoolPath + "/" + ImageName + " -outfile " + 
			serverinfo.SigntoolPath + "/" + outfilename +
			" -attribute " + imageattr.TargetAddr + " -attribute " + imageattr.NoJump  +
			" -algo" + " aes256" + " -wrapkey " + wrapkey + " -cert " + cert + " -sm" + " -secure" +
			" -encattr" + " -manuboot" + " -cred " + featurecred +  " -keyblob " + otpkeyblob
	} else {
		args = "sign" + " -type" + " BLw" + " -prikey " + privkey + " -wrapkey " +
			wrapkey + " -cert " + cert + " -encattr" + " -kdk" + " -sm" + " -secure" +
			" -pubkeytype" + " otp" + " -algo" + " aes256" + " -attribute " +
			imageattr.OsEntryAddr + " -attribute " + imageattr.TargetAddr + " -attribute " +
			imageattr.NoJump + " -attribute " + imageattr.Rollback + " -infile " +
			serverinfo.SigntoolPath + "/" + ImageName + " -outfile " +
			serverinfo.SigntoolPath + "/" + outfilename
	}
	fmt.Println("sign command: " + signtoolpath + " " + args)
	cmd := exec.Command("sshpass", "-p", Passwrd, "ssh",srvinfo, signtoolpath , args)
	var stderr bytes.Buffer
        cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		fmt.Println("Trying again ...")
		cmd := exec.Command("sshpass", "-p", Passwrd, "ssh",srvinfo, signtoolpath , args)
		err := cmd.Run()
		cmd.Stderr = &stderr
		if err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
			return fmt.Errorf("Imagesigning : SSH Image signing command failed: %w", err)
		}
	}

	return nil
}

func SignImage( serverinfo types.Server, ImagePath string, imghdr types.ImageHdrInfo,
		Passwrd string ) error {
	//copy image to server
	err := CopyImagetoServer( serverinfo, ImagePath, Passwrd )
	if err != nil {
		fmt.Println(err)
		return err
	}
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("SignImage : Getting working directory failed: %w", err)
	}
	curdir := filepath.Dir(exe)
        //Reading server configuration
        imgattrjson := curdir + "/image_attr.json"

	//Get Images configuration
	var images types.Images
        jFile, err := os.Open(imgattrjson)
        if err != nil {
                fmt.Println(err)
		return err
        }
        defer jFile.Close()
        imagesArray, _ := ioutil.ReadAll(jFile)
        json.Unmarshal(imagesArray, &images)

	//invoke ssh signing
	ss := strings.Split( ImagePath , "/" )
	ImgName := ss[len(ss)-1]

	for i := 0; i < len(images.Images); i++ {
		if ( images.Images[i].Type == imghdr.ImageType ){
                        fmt.Println("Invoking ssh image signing for " + imghdr.ImageType + " image")
                        Imagesigning(ImgName,images.Images[i],serverinfo,Passwrd)
			if err != nil {
				fmt.Println(err)
				return err
			}
                }
        }
	//copy image from server
	err = CopyImageFromServer( serverinfo, ImagePath, Passwrd )
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

