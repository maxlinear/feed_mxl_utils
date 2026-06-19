#
# Copyright (c) 2020-2022, MaxLinear, Inc.
#
# UGW Image Assembler Utility Host Make.
# This provides signing utility under staging_dir/host/bin/
#

include $(TOPDIR)/rules.mk
include $(INCLUDE_DIR)/kernel.mk
include $(INCLUDE_DIR)/version.mk

PKG_NAME:=imgassembler
PKG_VERSION:=1.0.0.1
PKG_RELEASE:=1

PKG_MAINTAINER:=MaxLinear
PKG_LICENSE:=MaxLinear
PKG_MXL_LICENSE:=MaxLinear/LICENSE

include $(INCLUDE_DIR)/package.mk
include $(INCLUDE_DIR)/host-build.mk

define Package/$(PKG_NAME)
  SUBMENU:=UGW Utilities
  SECTION:=utils
  CATEGORY:=MaxLinear
  TITLE:=UGW image assembler
  URL:= http://www.maxlinear.com
  MAINTAINER:=MaxLinear Inc
  HOST_BUILD_DEPENDS:=$(PKG_NAME)/host
  HOST_BUILD_DEPENDS:=golang/host
  HOST_BUILD_DEPENDS:=openssl
  PKG_BUILD_DEPENDS:=$(PKG_NAME)/host
endef

define Package/$(PKG_NAME)/description
 UGW image assembler  
endef

define Host/Prepare
	@echo  $(HOST_BUILD_DIR)
	$(INSTALL_DIR) $(HOST_BUILD_DIR)/imgassembler
	$(CP) -L ./src/* $(HOST_BUILD_DIR)/imgassembler
endef

define Host/Compile
	@echo  $(HOST_BUILD_DIR)
	@cd $(HOST_BUILD_DIR)/imgassembler;$(STAGING_DIR_HOSTPKG)/bin/go mod init imgassembler; $(STAGING_DIR_HOSTPKG)/bin/go build
	$(HOSTCC) $(HOST_BUILD_DIR)/imgassembler/openssl_test_util/genSBCR_openssl.c -lcrypto \
			-o $(HOST_BUILD_DIR)/imgassembler/openssl_test_util/genSBCR_openssl
endef

define Host/Install
	$(INSTALL_BIN) $(HOST_BUILD_DIR)/imgassembler/imgassembler $(STAGING_DIR_HOST)/bin/
endef

$(eval $(call HostBuild))
$(eval $(call BuildPackage,$(PKG_NAME)))
