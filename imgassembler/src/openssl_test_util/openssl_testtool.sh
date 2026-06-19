#!/bin/sh
# Copyright (c) 2020-2022, MaxLinear, Inc.
# This script genertaes signature, IEK and iv
# Encrypts image and header 
# Generates sbcr wrapped key 
# $1 - image to be encrypted
# $2 - private key
# $3 - huk

imagebinfile=$1
privkeybinfile=$2
hukbinfile=$3
OPENSSL="/home/qateam/openssl/openssl-3.0.1/apps/openssl"

rm -f signature.bin
rm -f sbcr.bin
rm -f enc_key.bin
rm -f iv.bin

#### Signature generation
$OPENSSL dgst -sha384 -sign $privkeybinfile -keyform DER -out sign.sha384 $imagebinfile

### parsing signature and getting r and s and forming signature
$OPENSSL asn1parse -inform DER -in sign.sha384 | sed 1d > sig.out
while read index
do
  echo "line: " $index
  output=$(printf "%s\n" "${index#*INTEGER }")
  output1=$(echo $output | cut -d ":" -f2)
  tmpsig="${tmpsig}${output1}"
done < sig.out
rm -f sig.out
sig=$tmpsig
echo "======= signature ============="
echo "signature: $sig"
echo "$sig" > sig_temp.bin
xxd -r -p > signature.bin < sig_temp.bin
rm -f sig_temp.bin
rm -f sign.sha384

#### Storing HUK
huk=$(xxd -p $hukbinfile | tr -d '\n')
echo "HUK key : $huk"

#### SBCR generation
sig_64byte=$(echo "${sig}" | awk '{print substr ($0, 0, 128)}')
echo "signature 64 bytes $sig_64byte"
sbcr=$( ./openssl_test_util/genSBCR $sig_64byte $huk)
echo "sbcr $sbcr"
echo "$sbcr" > sbcr_temp.bin
xxd -r -p > sbcr.bin < sbcr_temp.bin
rm -f sbcr_temp.bin

#### generating IEK and iv
output="$($OPENSSL enc -nosalt -aes-256-cbc -k mypass -P)"
echo "$output"
output1=$( echo "$output" | grep "key")
IEK=$(echo "$output1" | awk '{ print substr ($0, 5 ) }')
echo "IEK:$IEK"
echo "$IEK" > iek_temp.bin
xxd -r -p > enc_key.bin < iek_temp.bin
rm -f iek_temp.bin

output1=$( echo "$output" | grep "iv")
iv=$(echo "$output1" | awk '{ print substr ($0, 5 ) }')
echo "iV:$iv"
echo "$iv" > iv_temp.bin
xxd -r -p > iv.bin < iv_temp.bin
rm -f iv_temp.bin

#### split image and attribute header to encrypt
dd if=$imagebinfile of=attr_hdr.bin bs=128 count=1
dd if=$imagebinfile of=image_wo_atthdr.bin bs=1 skip=128

####  encrypting image and header
$OPENSSL enc -nopad -nosalt -aes-256-cbc -in image_wo_atthdr.bin -out image_wo_atthdr.bin.enc -K $IEK -iv $iv
$OPENSSL enc -nopad -nosalt -aes-256-cbc -in attr_hdr.bin -out attr_hdr.bin.enc -K $IEK -iv $iv

#### Creating encrypted image + header
cat attr_hdr.bin.enc image_wo_atthdr.bin.enc > $imagebinfile.enc

rm -f attr_hdr.bin 
rm -f image_wo_atthdr.bin
rm -f attr_hdr.bin.enc 
rm -f image_wo_atthdr.bin.enc 

#### encrypting the IEK using sbcr
$OPENSSL enc -iv A6A6A6A6A6A6A6A6 -K $sbcr -id-aes256-wrap -in enc_key.bin -out enc_key.bin.wrap
