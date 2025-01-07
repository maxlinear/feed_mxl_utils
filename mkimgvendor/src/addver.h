/*
 * Copyright (c) 2019 Intel Corporation
 * Copyright (c) 2020 MaxLinear Inc
 *
 * Tool to add vendor related options part of image.
 *
 */

#ifndef _VENDOR_HDR_H
#define _VENDOR_HDR_H

#define MAX_VEN_LEN 64

typedef struct vendor {
        char model[MAX_VEN_LEN]; /* model name */
	/*more vendor option can be supported by extending this structure */
}xvendor_t;

#endif
