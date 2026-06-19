#!/bin/sh                                                                                                                                                                                 
# Wrapper script to generate secureboot image using image assembler
# $1 - image asembler path
# $2 - openssl script path
# $3 - private key
# $4 - huk
# $5 - certificate
# $6 - tep to be encrypted
# $7 - bootloader to be encrypted
# $8 - rbei to be encrypted



imgassembler=$1
openssl_script=$2
privkeybinfile=$3
hukbinfile=$4
certificate=$5

tep=$6
uboot=$7
rbe=$8


echo "***************** signning tep $tep *****************"
$imgassembler ah attribute image_attr_tep.json image $tep
$openssl_script  $tep.trim.with_attr_hdr $privkeybinfile $hukbinfile
$imgassembler sbimage enc-image $tep.trim.with_attr_hdr.enc signature signature.bin iv iv.bin key enc_key.bin.wrap cert-count 1 certificate $certificate attribute image_attr_tep.json out-image $tep.signed


echo "***************** signning uboot $uboot *****************"
$imgassembler ah attribute image_attr_uboot.json image $uboot
$openssl_script  $uboot.trim.with_attr_hdr $privkeybinfile $hukbinfile
$imgassembler sbimage enc-image $uboot.trim.with_attr_hdr.enc signature signature.bin iv iv.bin key enc_key.bin.wrap cert-count 1 certificate $certificate attribute image_attr_uboot.json out-image $uboot.signed

echo "***************** signning rbe $rbe *****************"
$imgassembler ah attribute image_attr_rbe.json image $rbe
$openssl_script  $rbe.trim.with_attr_hdr $privkeybinfile $hukbinfile
$imgassembler sbimage enc-image $rbe.trim.with_attr_hdr.enc signature signature.bin iv iv.bin key enc_key.bin.wrap cert-count 1 certificate $certificate attribute image_attr_rbe.json out-image $rbe.signed rbe-orig $rbe

