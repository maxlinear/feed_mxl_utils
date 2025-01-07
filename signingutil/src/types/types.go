/*
 * Copyright (c) 2020-2022, MaxLinear, Inc.
 *
 * This file has all the data types defined
 *
 */

package types


/* Servers struct which contains an array of servers*/
type Servers struct {
        Servers []Server `json:"servers"`
}

/* Server struct which contains ip, username, password and type field*/
type Server struct {
        Host string `json:"host")`
        Type string `json:"type")`
        UserName string `json:"username")`
        Password string `json:"password")`
        SigntoolPath string `json:"signtool_path")`
        PrivKey string `json:"privkey")`
        WrapKey string `json:"wrapkey")`
        Cert string `json:"cert")`
        FeatureCred string `json:"feature_cred")`
        OtpKeyBlob string `json:"otp_key_blob")`
}

/* Images struct which contains an array of ImgAttr*/
type Images struct {
        Images []ImageAttr `json:"images"`
}

/* Server struct which contains ip, username, password and type field*/
type ImageAttr struct {
        Type string `json:"type")`
        OsEntryAddr string `json:"osentryaddr")`
        TargetAddr string `json:"targetaddr")`
        NoJump string `json:"nojump")`
        Rollback string `json:"rollback")`
        ImgId string `json:"imgid")`
}

/* Images cfg struct which contains an array of img cfg*/
type ImagesCfg struct {
        FullImages []FullimgCfg `json:"imgscfg"`
}

/* Server struct which contains ip, username, password and type field*/
type  FullimgCfg struct {
        DtbName string `json:"dtb")`
        FullImgName string `json:"fullimage")`
}


type ImageHdrInfo struct {
	ImageName	string
	ImageType	string
	LoadAddr	string
	Entrypoint	string
	Compression	string
}

