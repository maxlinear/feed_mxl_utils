#
# Copyright (C) 2006-2009 OpenWrt.org
#
# This is free software, licensed under the GNU General Public License v2.
# See /LICENSE for more information.
#

include $(TOPDIR)/rules.mk
include $(INCLUDE_DIR)/kernel.mk

PKG_NAME:=cpuload
PKG_VERSION:=1.0.0
PKG_RELEASE:=1

include $(INCLUDE_DIR)/package.mk

define KernelPackage/cpuload
 SUBMENU:=MaxLinear
 TITLE:=Kernel support for per-VPE cpuload for MIPS
 DEPENDS:=@TARGET_intel_mips +@TARGET_intel_mips_xrx500
 FILES:=$(PKG_BUILD_DIR)/cpuload.ko

 AUTOLOAD:=$(call AutoProbe,cpuload)
endef

define KernelPackage/cpuload/description
 Kernel support for per-VPE cpuload tool for MIPS
endef

define Build/Compile
	$(KERNEL_MAKE) SUBDIRS="$(PKG_BUILD_DIR)" modules
endef

$(eval $(call KernelPackage,$(PKG_NAME)))
