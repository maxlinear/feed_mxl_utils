#
# Copyright (C) 2015 OpenWrt.org
#
# This is free software, licensed under the GNU General Public License v2.
# See /LICENSE for more information.
#

include $(TOPDIR)/rules.mk
include $(INCLUDE_DIR)/kernel.mk
include $(INCLUDE_DIR)/version.mk

PKG_NAME:=mkimgvendor
PKG_VERSION:=1.0.0
PKG_RELEASE:=1

PKG_MAINTAINER:=MaxLinear
PKG_LICENSE:=MaxLinear
PKG_MXL_LICENSE:=MaxLinear/LICENSE

PKG_BUILD_DEPENDS:=$(PKG_NAME)/host

include $(INCLUDE_DIR)/host-build.mk
include $(INCLUDE_DIR)/package.mk
#include $(INCLUDE_DIR)/cmake.mk

define Package/$(PKG_NAME)
  SUBMENU:=UGW Utilities
  SECTION:=utils
  CATEGORY:=MaxLinear
  TITLE:=UGW utility to add header
  URL:= http://www.maxlinear.com
endef

define Package/$(PKG_NAME)/description
    UGW utility to help to add header
endef

#CMAKE_OPTIONS += \

#CMAKE_HOST_OPTIONS += \

define Host/Prepare
	$(CP) -L ./src/* $(HOST_BUILD_DIR)/
endef

define Build/InstallDev
	$(INSTALL_DIR) $(1)/usr/include
	$(INSTALL_BIN) $(HOST_BUILD_DIR)/*.h $(1)/usr/include/
endef

define Host/Install
	$(INSTALL_BIN) $(HOST_BUILD_DIR)/mkimgvendor $(STAGING_DIR_HOST)/bin/
endef

$(eval $(call HostBuild))
$(eval $(call BuildPackage,$(PKG_NAME)))
