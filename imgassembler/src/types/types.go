/*
 * Copyright (c) 2020-2022, MaxLinear, Inc.
 *
 * This file has all the data types defined
 *
 */

package types

type AttributeElement struct {
	ElementType uint32
	ElementValue uint32
}

type SBImage struct {
	Type string
	Version uint8
	PubKeyType uint32
	Signature string
	EncryptionKey string
	IV string
	ImageLen uint32
	CertCount uint32
	Certificate string
}

/* Images struct which contains an array of ImgAttr*/
type ImgsAttrs struct {
	ImagesArray []ImgAttr `json:"attributes"`
}
type ImgAttr struct {
	TypeId string `json:"typeid")`
	AttrVal string `json:"value")`
}
