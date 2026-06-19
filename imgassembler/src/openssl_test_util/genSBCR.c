/*
 * Copyright (c) 2020-2022, MaxLinear, Inc.
 * 
 * API to generate sbcr
 *
 */

#include <stdio.h>
#include <string.h>
#include "openssl/ossl_typ.h"
#include "openssl/hmac.h"
#include "openssl/app_libctx.h"
#include "openssl/evp.h"
#define LABEL_LEN 64
#define HUK_LEN 32
#define CONTEXT_LEN 14
#define EVP_MAX_MD_SIZE 64

void gen_sbcr(char *huk, char *signature, unsigned char *sbcr);

int main(int argc, char **argv){
	int nRet = 0;
	if (argc < 3) {
		fprintf(stderr, "Specify proper arguments!\n");
		return 1;
	}
	char *sig = argv[1];
	char *huk = argv[2];
	unsigned char buf[EVP_MAX_MD_SIZE] = {0};
	nRet = gen_sbcr(huk,sig,buf);
	if (nRet != 0) {
		return 1;
	}
	return 0;
}
int gen_sbcr(char *huk, char *signature, unsigned char *sbcr)
{
	unsigned char label[LABEL_LEN] = {0};
	unsigned char context[CONTEXT_LEN] = {0x00,0x01,0x00,0x00,0x00,0x20,0x20,0x00,0x00,0x00,0xA9,0x20,0x00,0x01};
	unsigned char hukarray[HUK_LEN]= {0x00};
	EVP_MAC *mac = NULL;
	size_t len = 0;
	EVP_MAC_CTX *ctx = NULL;
	OSSL_PARAM params[2];
	int nIndex = 0;
	for (nIndex = 0; nIndex < HUK_LEN ; nIndex++) {
		sscanf(huk + 2*nIndex, "%02hhx", &hukarray[nIndex]);
		//printf("bytearrayi huk %d: %02hhx ", nIndex, hukarray[nIndex]);
	}
	for (nIndex = 0; nIndex < LABEL_LEN ; nIndex++) {
		sscanf(signature + 2*nIndex, "%02hhx", &label[nIndex]);
		//printf("bytearray sig %d: %02hhx ", nIndex, label[nIndex]);
	}
	mac = EVP_MAC_fetch(NULL, "HMAC" , NULL);
	if (mac == NULL) {
		printf("\nmac error");
		return 1;
	}

	ctx = EVP_MAC_CTX_new(mac);
	if (ctx == NULL) {
		printf("\nctx error");
		return 1;
	}
	params[0] = OSSL_PARAM_construct_utf8_string("digest", "SHA256", 0);
	params[1] = OSSL_PARAM_construct_end();
	EVP_MAC_CTX_set_params(ctx, params);
	if (!EVP_MAC_init(ctx, hukarray, HUK_LEN, params)) {
		printf("\n mac init error");
		return;
	}
	
	EVP_MAC_update(ctx, (unsigned char*)label, LABEL_LEN);
	EVP_MAC_update(ctx, (unsigned char*)context, CONTEXT_LEN);
	EVP_MAC_final(ctx, sbcr, NULL, 32);
	EVP_MAC_CTX_free(ctx);
	//printf("\nlen %lu",len);
	for (nIndex = 0; nIndex < 32; nIndex++) {
		printf("%02x", sbcr[nIndex]);
	}
	return 0;
}


