/*
 * Copyright (c) 2020-2022, MaxLinear, Inc.
 * 
 * Tool to sign and create fullimage
 *
 */
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"signingutil/cmnApi"
)

func main() {

	num_args := len(os.Args)
	flag := 0
	fmt.Println(num_args);
	if num_args < 5 {
		fmt.Println("Please specify proper arguments")
		fmt.Println("./signingUtil [operation] [Image Path] [username][password] <imgcfg.json>")
		fmt.Println("Ex: ./signingUtil sign /local/lgm_images qateam qateam /tmp/imgcfg.json")
		fmt.Println("Ex: ./signingUtil rotsign [manubootimg] [custimage] [fieldimage] qateam qateam")
		fmt.Println("Ex: ./signingUtil totalimg_up uboot [uboot fullpath] tepfw [tepfw fullpath] dtb [dtb fullpath]" +
				"kernel [kernel fullpath] rootfs [rootfs fullpath] output [outputimage fullpath")
		return
	}

	arguments := os.Args[1:]
	for index, arg := range arguments {
		switch {
			case arg == "sign":
				fmt.Println("Sign operation invoked")
				index += 2
				fInfo, err := os.Stat(os.Args[index])
				if err != nil {
					fmt.Println("Error in accesing path, please provide proper path")
					return
				}
				i := index + 1
				username := os.Args[i]
				fmt.Println("username : " + username)
				i = i + 1
				passwrd := os.Args[i]
				var imgcfgjson string
				if  num_args > 5 {
					i = i + 1
					imgcfgjson = os.Args[i]
				}
				if fInfo.IsDir() {
					files := []string{}
					err := filepath.Walk(os.Args[index], func(path string, f os.FileInfo, err error) error {
						if !f.IsDir() {
							files = append(files, path)
						}
						return err
					})
					//read each file and invoke signing
					for _, file := range files {
						err = cmnApi.Sign( file, username, passwrd )
						if err != nil {
							fmt.Println("Signing failed for the image" + file)
						}
					}
					//create fullimage
					err = cmnApi.CreateFullimage(os.Args[index],imgcfgjson)
					if err != nil {
						fmt.Println("Creating Fullimage failed");
						return
					}
				} else {
					cmnApi.Sign(os.Args[index] ,username, passwrd)
					if err != nil {
						return
					}
				}
				break
			case arg == "rotsign":
				fmt.Println("ROT Sign operation invoked")
				if num_args < 7 {
					fmt.Println("Please specify proper arguments")
					fmt.Println("./signingUtil [operation] [manubootimg] [custimage] [fieldimage] [username][password]")
					return
				}
				index += 2
				manuboot := os.Args[index]
				custimage := os.Args[index+1]
				fieldimage := os.Args[index+2]
				username := os.Args[index+3]
				password := os.Args[index+4]
				err := cmnApi.ROTSign( manuboot, username, password )
				if err != nil {
					fmt.Println("Signing failed for the image" + manuboot);
					return;
				}
				err = cmnApi.CreateROTFullimage(manuboot,custimage,fieldimage)
				if err != nil {
					fmt.Println("Creating ROT fullimage failed" + manuboot);
					return;
				}
				fmt.Println(manuboot)
				fmt.Println(custimage)
				fmt.Println(fieldimage)
				fmt.Println(username)
				fmt.Println(password)
				break
			case arg == "totalimg_up":
				var ubootImg string
				var tepfwImg string
				var dtbImg string
				var kernelImg string
				var rootfsImg string
				var outputImg string
				var err error
				fmt.Println("total image creation for user data partition is invoked")
				if num_args < 14  {
					fmt.Println("Please specify proper arguments")
					fmt.Println("Ex: ./signingUtil totalimg_up uboot [uboot fullpath] tepfw [tepfw fullpath] dtb [dtb fullpath]" +
						"kernel [kernel fullpath] rootfs [rootfs fullpath] output [outputimage fullpath")
					return
				}
				for i, a := range os.Args[2:] {
					switch {
						case a == "uboot":
							ubootImg = os.Args[i+3]
							fmt.Println("ubootImg :" + ubootImg)
						break
						case a == "tepfw":
							tepfwImg = os.Args[i+3]
							fmt.Println("tepfwimg :" + tepfwImg)
						break
						case a == "dtb":
							dtbImg = os.Args[i+3]
							fmt.Println("dtbimg :" + dtbImg)
						break
						case a == "kernel":
							kernelImg = os.Args[i+3]
							fmt.Println("kernelimg :" + kernelImg)
						break
						case a == "rootfs":
							rootfsImg = os.Args[i+3]
							fmt.Println("rootfsImg :" + rootfsImg)
						break
						case a == "output":
							outputImg = os.Args[i+3]
							fmt.Println("outputImg :" + outputImg)
						break

					}
				}
				err = cmnApi.CreateUpTotalImg(ubootImg,tepfwImg,dtbImg,kernelImg,rootfsImg,outputImg)
				if err != nil {
					fmt.Println(err)
					return
				}

				flag = 1
				break
			case arg == "help":
				fmt.Println("./signingUtil [operation] [Image Path] [username] [password] <imgcfg.json>")
				fmt.Println("Ex: ./signingUtil sign /local/lgm_images qateam qateam /tmp/imgcfg.json")
				fmt.Println("Ex: ./signingUtil rotsign [manubootimg] [custimage] [fieldimage] qateam qateam")
				break
		}
		if flag == 1 {
			break
		}
	}
}

