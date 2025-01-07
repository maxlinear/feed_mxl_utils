#include <stdio.h>
#include <string.h>
#include "openssl/ossl_typ.h"
#include "openssl/hmac.h"
#define LABEL_LEN 64
#define HUK_LEN 32
#define CONTEXT_LEN 14
#define EVP_MAX_MD_SIZE 64
#define HMAC_RETURN_SUCCESS 1
int gen_sbcr(unsigned char *huk, unsigned char *signature, unsigned char *sbcr);
int main(int argc, char **argv) {
	int nRet = 0;
	if(argc < 3){
		printf("Specify proper arguments!\n");
		printf("Example:genSBCR_openssl <signature> <huk>\n");
		return 1;
	}
	char *sig = argv[1];
	char *huk = argv[2];
	unsigned char buf[EVP_MAX_MD_SIZE] = {0};
	int nHuklen = strlen(huk);
	int nSiglen = strlen(sig);
	if(nHuklen != (HUK_LEN * 2)) {
		printf("HUK size should be 32 byte\n");
		return 1;
	}
	if(nSiglen != (LABEL_LEN*2)) {
		printf("signature size should be 64 byte\n");
		return 1;
	}
	nRet = gen_sbcr(huk, sig, buf);
	if (nRet != 0) {
		return 1;
	}
	return 0;
}
int gen_sbcr(unsigned char *huk, unsigned char *signature, unsigned char *sbcr) {
	unsigned char label[LABEL_LEN] = {0};
	unsigned char context[CONTEXT_LEN] = {0x00,0x01,0x00,0x00,0x00,0x20,0x20,0x00,0x00,0x00,0xA9,0x20,0x00,0x01};
	unsigned char hukarray[HUK_LEN]= {0x00};
	int nIndex = 0;	
	int nSbcrlen = 0;
	int nRet = HMAC_RETURN_SUCCESS;
	for (nIndex = 0; nIndex < HUK_LEN ; nIndex++) {
		sscanf(huk + 2*nIndex, "%02hhx", &hukarray[nIndex]);
		//printf("bytearrayi huk %d: %02hhx ", nIndex, hukarray[nIndex]);
	}
	for (nIndex = 0; nIndex < LABEL_LEN ; nIndex++) {
		sscanf(signature + 2*nIndex, "%02hhx", &label[nIndex]);
		//printf("bytearray sig %d: %02hhx ", nIndex, label[nIndex]);
	}

	HMAC_CTX *ctx = NULL;
	ctx = HMAC_CTX_new();
	if (ctx == NULL) {
		printf("\nctx error");
		return 1;
	}
	HMAC_CTX_reset(ctx);

	nRet = HMAC_Init_ex(ctx, hukarray, HUK_LEN, EVP_sha256(), NULL);
	if(nRet != HMAC_RETURN_SUCCESS){
		printf("openssl api HMAC_Init_ex failed to update digest\n");
		return 1;
	}
	nRet = HMAC_Update(ctx, (unsigned char*)label, LABEL_LEN);
	if(nRet != HMAC_RETURN_SUCCESS){
		printf("openssl api HMAC_Update failed to update signature\n");
		return 1;
	}
	nRet = HMAC_Update(ctx, (unsigned char*)context, CONTEXT_LEN);
	if(nRet != HMAC_RETURN_SUCCESS){
		printf("openssl api HMAC_Update failed to update context\n");
		return 1;
	}
	
	nRet = HMAC_Final(ctx, sbcr, &nSbcrlen);
	if(nRet != HMAC_RETURN_SUCCESS){
		printf("openssl api HMAC_Final failed to generate sbcr\n");
		return 1;
	}
	for (int i = 0; i < nSbcrlen; i++) {
		printf("%02x", sbcr[i]);
	}
	printf("\n");
	HMAC_CTX_free(ctx);
	return 0;
}


