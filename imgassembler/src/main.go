/*
 * Copyright (c) 2020-2022, MaxLinear, Inc.
 * 
 * Tool to assemble  image
 *
 */
package main

import (
	"fmt"
	"os"
	"strconv"
	"imgassembler/attrHdrCreation"
	"imgassembler/sbImageCreation"
	"imgassembler/types"
)

func main() {

	num_args := len(os.Args)
	if num_args < 2  {
		fmt.Println("Please specify proper arguments")
		fmt.Println("imgassembler  command [-command-opts] [command-args]")
		fmt.Println("Ex: ./imgassembler ah attribute file.json image imagefile.bin")
		fmt.Println("Ex: ./imgassembler sbimage enc-image image-file-signed.bin" +
		            " signature sig.bin iv iv.bin key key.bin cert-count count certificate" +
		            " cert.bin attribute file.json out-image sec-img.bin rbe-orig rbe_orig.bin")
		return
	}

	arguments := os.Args[1:]
	recovery := 0
	for index, arg := range arguments {
		switch {
			case arg == "ah":
				fmt.Println("attribute header creation")
				if num_args < 6  {
					fmt.Println("Please specify proper arguments")
					fmt.Println("imgassembler  command [-command-opts] [command-args]")
					fmt.Println("Ex: ./imgassembler  ah attribute file.json image imagefile.bin")
					return
				}
				index += 3
				fmt.Println(os.Args[index])
				fmt.Println(os.Args[index+2])
				if num_args == 7 {
					fmt.Println(os.Args[index+3])
					if os.Args[index+3] == "recovery" {
						recovery = 1
					}
				}
				err := attrHdrCreation.AttrHdr(os.Args[index] , os.Args[index+2] , recovery)
				if err != nil {
					fmt.Println(err)
					return
				}
				break
			case arg == "sbimage":
				fmt.Println("Secure boot image creation invoked")
				if num_args < 17  {
					fmt.Println("Please specify proper arguments")
					fmt.Println("imgassembler  command [-command-opts] [command-args]")
					fmt.Println("Ex: ./imgassembler sbimage enc-image image-file-signed.bin" +
						" signature sig.bin iv iv.bin key key.bin cert-count count certificate" +
						" cert.bin attribute file.json out-image sec-img.bin rbe-orig rbe_orig.bin")
					return
				}
				var sbImg types.SBImage
				var outImgPath string
				var inputImgPath string
				var origImgPath string
				var attrFile string
				var err error
				var val int
				for i, a := range os.Args[2:] {
					switch {
						case a == "enc-image":
							inputImgPath = os.Args[i+3]
							fmt.Println("inputImgPath :" + inputImgPath);
						break
						case a == "signature":
							sbImg.Signature = os.Args[i+3]
							fmt.Println("signature :" + sbImg.Signature);
						break
						case a == "iv":
							sbImg.IV = os.Args[i+3]
							fmt.Println("iv :" + sbImg.IV);
						break
						case a == "key":
							sbImg.EncryptionKey = os.Args[i+3]
							fmt.Println("key :" + sbImg.EncryptionKey);
						break
						case a == "cert-count":
							val, err = strconv.Atoi(os.Args[i+3])
							if err != nil {
								fmt.Println("strconv.Atoi failed for " + os.Args[i+3])
							}
							sbImg.CertCount = uint32(val)
							fmt.Println(sbImg.CertCount);
						break
						case a == "certificate":
							sbImg.Certificate = os.Args[i+3]
							fmt.Println("Certificate :" + sbImg.Certificate);
						break
						case a == "attribute":
							attrFile = os.Args[i+3]
							fmt.Println("attrFile" + attrFile);
						break
						case a == "out-image":
							outImgPath = os.Args[i+3]
							fmt.Println("outImgPath " + outImgPath);
						break
						case a == "rbe-orig":
							origImgPath = os.Args[i+3]
							fmt.Println("origImgPath " + origImgPath);
						break
						case a == "recovery":
							recovery = 1
						break
					}
				}
				err = sbImageCreation.CreateSBImage(inputImgPath,origImgPath,outImgPath,sbImg,attrFile,recovery)
				if err != nil {
					fmt.Println(err)
					return
				}
				break
			case arg == "help":
				fmt.Println("imgassembler  command [-command-opts] [command-args]")
				fmt.Println("Ex: ./imgassembler  ah attribute file.json image imagefile.bin")
				fmt.Println("Ex: ./imgassembler sbimage enc-image image-file-signed.bin" +
				            " signature sig.bin iv iv.bin key key.bin cert-count count certificate" +
				            " cert.bin attribute file.json out-image sec-img.bin rbe-orig rbe_orig.bin")
				break
		}
	}
}

