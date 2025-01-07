/*
 * Copyright (c) 2019 Intel Corporation
 * Copyright (c) 2020 MaxLinear Inc
 * 
 * Tool to add vendor related options part of image.
 *
 */
#include <stdio.h>
#include <string.h>
#include <unistd.h>
#include "addver.h"

int main(int argc, char **argv) {

	char *path = NULL, *model_name = NULL;
	FILE *file=NULL;
	xvendor_t hdr = {0};
	int opt;
	while((opt = getopt(argc, argv, ":p:m:s:o:")) != -1)
	{
		switch(opt)
		{
			case 'p':
				printf("path name :%s\n", optarg);
				path = optarg;
				break;
			case 'm':
				printf("modelname: %s\n", optarg);
				model_name = optarg;
				snprintf(hdr.model,MAX_VEN_LEN,"%s",model_name);;
        			
				break;
			case 's':
			case 'o':
				break;
			case '?':
				printf("unknown option: %c\n", optopt);
				break;
		}
	}

	if (path != NULL && model_name != NULL) {
		file= fopen(path, "wb");
		if (file != NULL) {
			fwrite(&hdr, sizeof(xvendor_t), 1, file);
			fclose(file);
		}
	}

	for(; optind < argc; optind++){
		printf("extra arguments: %s\n", argv[optind]);
	}

	return 0;
}
