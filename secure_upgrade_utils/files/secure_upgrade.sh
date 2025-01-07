#!/bin/sh

#secure upgrade script to perform upgrade operation from procd.
. /etc/profile.d/intel.sh


UPG_UTILTIY="/opt/intel/usr/sbin/secupg"
IMG_PATH="/tmp/upgrade/firmware.img"
SEC_IMG_PATH="/tmp/secupg/firmware.img"

#default upgrade operation later more operations will be added
UPG=1
REBOOT=1

UPG_NONE=0
UPG_REQ=1
UPG_INPROG=2
UPG_SUCC=3
UPG_FAIL=4
STATE_PATH="/opt/intel/etc/upgrade/.upg_state_file"

c_clr="\033[00;00m"
c_gre="\033[32;01m"
c_red="\033[31;01m"

_info()
{
        echo -en "\n${c_gre}$@${c_clr}\n";
}

_hlptxt()
{
        local _opt="$1"
        shift
        echo -en "      $c_red${_opt}${c_clr}  $@\n"
}


_help()
{
        _hlptxt "Secure system upgarde ."
        _hlptxt "Usage: $0 [ options ]";
        _hlptxt "     $0 -u <img path> "
        _hlptxt "     $0 -r <img path> (no reboot)"
        exit $1;
}

[ -n "$1" ] && {
        case "$1" in
                -h) _help 0;
                ;;
                -u) UPG=1;
                ;;
                -r) REBOOT=0; UPG=1;
                ;;
                *) _help 1;
        esac
}

[ ! -d "/mnt/upgrade" ] && mkdir -p /tmp/mount_upgrade && ln -s /tmp/mount_upgrade /mnt/upgrade

[ -n "$UPG" ] && {
        logger -t "secupg secupg" "SYSTEM UPGRADE Start"
	if [ "$REBOOT" != "0" ] ; then
		mkdir -p /tmp/secupg/
		chmod 700 /tmp/secupg/
		mv $IMG_PATH $SEC_IMG_PATH
		$UPG_UTILTIY -u $SEC_IMG_PATH
		_ret=$?
		if [ $_ret -eq 0 ] ; then
			logger -t "secupg secupg" "SYSTEM UPGRADE Completed... Going for Reboot "
			sync
			ubus call system reboot &
			exit
		else
			if [ $_ret -eq 1 ] ; then
				logger -t "secupg secupg" "SYSTEM UPGRADE Validation Failed... Abort UPGRADE... "
				echo $UPG_FAIL > $STATE_PATH
				return 1;
			else
				logger -t "secupg secupg" "SYSTEM UPGRADE Flash Write Failed... Going for Reboot... "
				echo $UPG_FAIL > $STATE_PATH
				ubus call system reboot &
				exit
			fi
		fi
	else
		$UPG_UTILTIY -r -u $IMG_PATH
		_ret=$?
		if [ $_ret -eq 0 ] ; then
			logger -t "secupg secupg" "SYSTEM UPGRADE Completed..."
			return 0
		else
			if [ $_ret -eq 1 ] ; then
				logger -t "secupg secupg" "SYSTEM UPGRADE Validation Failed... Abort UPGRADE... "
				echo $UPG_FAIL > $STATE_PATH
				return $_ret
			else
				logger -t "secupg secupg" "SYSTEM UPGRADE Flash Write Failed... Reboot recommended... "
				echo $UPG_FAIL > $STATE_PATH
				return $_ret
			fi
		fi
	fi
}
