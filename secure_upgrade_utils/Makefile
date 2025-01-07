#
# Copyright (C) 2009 OpenWrt.org
# Copyright (C) 2019 Intel Corporation
# Copyright © 2020-2024 MaxLinear, Inc.
#
# Openwrt Makefile for upgrade
#
#

#### Includes ###################################
include $(TOPDIR)/rules.mk
include $(INCLUDE_DIR)/kernel.mk

#### Package Info ############################### 
PKG_NAME:=secure_upgrade_utils
PKG_SOURCE_VERSION:=2.2.1.17
PKG_SOURCE_PROTO:=git
CONFIG_UGW_PKG_SOURCE_URL:=ssh://git@git.maxlinear.com
PKG_PROJECT:=ugw
PKG_SOURCE_NAME:=upgrade_utils
PKG_SOURCE_INNERDIR:=secure_upgrade
PKG_SOURCE_URL:=$(CONFIG_UGW_PKG_SOURCE_URL)/$(PKG_PROJECT)/$(PKG_SOURCE_NAME).git
PKG_MIRROR_HASH:=dc59ce3c54d35987e02873cf79621a04c7757b1faa6f9363f206fcd5421e5552

PKG_BUILD_DIR:=$(BUILD_DIR)/$(PKG_NAME)-$(PKG_SOURCE_VERSION)
PKG_MAINTAINER:=MaxLinear
PKG_LICENSE:=MaxLinear
PKG_MXL_LICENSE:=MaxLinear/LICENSE

include $(INCLUDE_DIR)/package.mk
-include $(INCLUDE_DIR)/package-version-override.mk
-include $(STAGING_DIR)/mk/ugw-loglevel.mk

OPENWRT_VERSION:=$(shell cat $(TOPDIR)/version | grep 22.03)

UGW_VERSION:=$(if $(wildcard $(TOPDIR)/ugw_version),$(shell cat $(TOPDIR)/ugw_version))
 
 PKG_DEPENDS:= \
         +TARGET_intel_mips:sse_fapi \
         +TARGET_intel_x86_lgm:sse_fapi_lgm \
         +libscapi

define Package/$(PKG_NAME)
  SECTION:=libs
  CATEGORY:=MaxLinear
  SUBMENU:=UGW Utilities
  TITLE:=UGW utility to support upgrade options
  URL:=http://www.maxlinear.com
  MAINTAINER:=MaxLinear Inc
  DEPENDS:= +libsafec3-7 +nand-utils $(PKG_DEPENDS)
  MENU:=1
endef

define Package/$(PKG_NAME)/description
	This package contains upgrade utilities for MaxLinear modules.
endef

### Menuconfig ##################################
define Package/$(PKG_NAME)/config
$(call Package/$(PKG_NAME)/override_version)
$(call Package/$(PKG_NAME)/override_source_path)
endef

TARGET_CFLAGS += -DHAVE_C99=1 \
	$(if $(CONFIG_TARGET_intel_mips),-DARCH_MIPS) \
	-DCONFIG_GCC_VERSION=\\\"$(CONFIG_GCC_VERSION)\\\" \
	-DCONFIG_LIBC_VERSION=\\\"$(CONFIG_LIBC_VERSION)\\\"

ifeq ($(CONFIG_TARGET_intel_x86_lgm),y)
	TARGET_CFLAGS += -DCONFIG_UBOOT_CONFIG_ENV_SIZE=$(CONFIG_INTEL_X86_ENV_SZ) \
	-DPLATFORM_LGM
else 
	TARGET_CFLAGS += -DCONFIG_UBOOT_CONFIG_ENV_SIZE=0x4000
ifeq ($(CONFIG_TARGET_intel_mips_prx300),y)
	MAKE_OPTS += PRX300=y
endif
endif
TARGET_CFLAGS += -I$(STAGING_DIR)/usr/include/safeclib

define Package/$(PKG_NAME)/install
	$(INSTALL_DIR) $(1)/usr/sbin
	$(INSTALL_DIR) $(1)/sbin
	$(INSTALL_DIR) $(1)/$(VENDOR_PATH)/usr/sbin
	$(INSTALL_DIR) $(1)/usr/share/capabilities/
	$(INSTALL_DIR) $(1)/usr/share/acl.d/
	
	$(INSTALL_BIN) $(PKG_BUILD_DIR)/read_img $(1)/usr/sbin/
	$(INSTALL_BIN) $(PKG_BUILD_DIR)/uboot_env $(1)/usr/sbin/
	$(INSTALL_BIN) $(PKG_BUILD_DIR)/upgrade $(1)/usr/sbin/
	$(INSTALL_BIN) $(PKG_BUILD_DIR)/secupg $(1)/$(VENDOR_PATH)/usr/sbin/
	$(INSTALL_DATA) ./files/upgrade_cap.json $(1)/usr/share/capabilities/
	$(if $(OPENWRT_VERSION), \
		$(INSTALL_DATA) ./files/upgrade_cap_owrt_22x.json $(1)/usr/share/capabilities/upgrade_cap.json
	)
	$(INSTALL_DATA) ./files/upgrade_acl.json $(1)/usr/share/acl.d/
	$(INSTALL_BIN) ./files/vol_mgmt $(1)/usr/sbin/
	$(INSTALL_BIN) ./files/secure_upgrade.sh $(1)/sbin/
ifeq ($(CONFIG_TARGET_intel_x86_lgm),y)
	$(INSTALL_BIN) ./files/uboot_lgm_img_vars.sh $(1)/sbin/uboot_img_vars.sh
else 
	$(INSTALL_BIN) ./files/uboot_img_vars.sh $(1)/sbin/
endif
endef

define Build/Clean
	if [ -d $(PKG_BUILD_DIR) ]; then $(MAKE) -C $(PKG_BUILD_DIR) clean; \
	$(RM) -r $(PKG_BUILD_DIR)/ipkg-$(BOARD); fi
endef

$(eval $(call BuildPackage,$(PKG_NAME)))
