#
# Copyright (c) 2020-2023, MaxLinear, Inc.
#
# UGW Signing Utility Host Make.
# This provides signing utility under staging_dir/host/bin/
#

include $(TOPDIR)/rules.mk
include $(INCLUDE_DIR)/kernel.mk
include $(INCLUDE_DIR)/version.mk

PKG_NAME:=signingutil
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
  TITLE:=UGW image signing utility
  URL:= http://www.maxlinear.com
  MAINTAINER:=MaxLinear Inc
  HOST_BUILD_DEPENDS:=$(PKG_NAME)/host
  HOST_BUILD_DEPENDS:=golang/host
  PKG_BUILD_DEPENDS:=$(PKG_NAME)/host
endef

define Package/$(PKG_NAME)/description
 UGW image signing utility  
endef

define Host/Prepare
	@echo  $(HOST_BUILD_DIR)
	$(INSTALL_DIR) $(HOST_BUILD_DIR)/signingutil
	$(CP) -L ./src/* $(HOST_BUILD_DIR)/signingutil
endef

define Host/Compile
	@echo  $(HOST_BUILD_DIR)
	@cd $(HOST_BUILD_DIR)/signingutil;$(STAGING_DIR_HOSTPKG)/bin/go mod init signingutil; $(STAGING_DIR_HOSTPKG)/bin/go build
endef

define Host/Install
	$(INSTALL_BIN) $(HOST_BUILD_DIR)/signingutil/signingutil $(STAGING_DIR_HOST)/bin/
	$(CP) $(HOST_BUILD_DIR)/signingutil/server.json	$(STAGING_DIR_HOST)/bin/
	$(CP) $(HOST_BUILD_DIR)/signingutil/image_attr.json $(STAGING_DIR_HOST)/bin/
	$(CP) $(HOST_BUILD_DIR)/signingutil/swap_bin.pl	$(STAGING_DIR_HOST)/bin/
	$(CP) $(HOST_BUILD_DIR)/signingutil/mk_intel_tail.pl $(STAGING_DIR_HOST)/bin/
	$(CP) $(HOST_BUILD_DIR)/signingutil/uart_bin2asc.pl $(STAGING_DIR_HOST)/bin/
	$(CP) $(HOST_BUILD_DIR)/signingutil/mk_intel_header.pl $(STAGING_DIR_HOST)/bin/
	$(CP) $(HOST_BUILD_DIR)/signingutil/pad2align.sh $(STAGING_DIR_HOST)/bin/
	$(CP) $(HOST_BUILD_DIR)/signingutil/hostcrc32 $(STAGING_DIR_HOST)/bin/
endef

$(eval $(call HostBuild))
$(eval $(call BuildPackage,$(PKG_NAME)))
