#!/bin/sh                                                                                                                                                                                 
# Wrapper script to generate secureboot image using image assembler
# $1 - image asembler path
# $2 - openssl script path
# $3 - private key
# $4 - huk
# $5 - certificate
# $6 - kernel image to be encrypted
# $7 - rootfs to be encrypted
# $8 - dtb to be encrypted


imgassembler=$1
openssl_script=$2
privkeybinfile=$3
hukbinfile=$4
certificate=$5

kernel=$6
rootfs=$7
dtb=$8
dtbo=$9

echo "***************** signning kernel $kernel *****************"
$imgassembler ah attribute image_attr_kernel.json image $kernel
$openssl_script  $kernel.trim.with_attr_hdr $privkeybinfile $hukbinfile
$imgassembler sbimage enc-image $kernel.trim.with_attr_hdr.enc signature signature.bin iv iv.bin key enc_key.bin.wrap cert-count 1 certificate $certificate attribute image_attr_kernel.json out-image $kernel.signed

echo "***************** signning rootfs $rootfs *****************"
$imgassembler ah attribute image_attr_rootfs.json image $rootfs
$openssl_script  $rootfs.trim.with_attr_hdr $privkeybinfile $hukbinfile
$imgassembler sbimage enc-image $rootfs.trim.with_attr_hdr.enc signature signature.bin iv iv.bin key enc_key.bin.wrap cert-count 1 certificate $certificate attribute image_attr_rootfs.json out-image $rootfs.signed

echo "***************** signning dtb $dtb *****************"
$imgassembler ah attribute image_attr_dtb.json image $dtb
$openssl_script  $dtb.with_attr_hdr $privkeybinfile $hukbinfile
$imgassembler sbimage enc-image $dtb.with_attr_hdr.enc signature signature.bin iv iv.bin key enc_key.bin.wrap cert-count 1 certificate $certificate attribute image_attr_dtb.json out-image $dtb.signed

echo "***************** signning dtbo $dtbo *****************"
$imgassembler ah attribute image_attr_dtbo.json image $dtbo
$openssl_script  $dtbo.with_attr_hdr $privkeybinfile $hukbinfile
$imgassembler sbimage enc-image $dtbo.with_attr_hdr.enc signature signature.bin iv iv.bin key enc_key.bin.wrap cert-count 1 certificate $certificate attribute image_attr_dtbo.json out-image $dtbo.signed


#echo "***************** create fullimage ********************"
#cat $kernel.signed $rootfs.signed $dtb.signed > fullimage.img.tmp
#mkimage -A x86 -O linux -T multi -a 0x00 -C none -e 0x00 -n 'LGM-fullimage' -d fullimage.img.tmp fullimage.img

