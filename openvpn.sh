#!/bin/bash

ABORT=0
cleanup() {
    ABORT=1
    echo "Fri Apr 12 03:40:09 2019 event_wait : Interrupted system call (code=4)"
    echo "Fri Apr 12 03:40:09 2019 Closing TUN/TAP interface"
    echo "Fri Apr 12 03:40:09 2019 /sbin/ifconfig tun0 0.0.0.0"
    echo "Fri Apr 12 03:40:09 2019 SIGINT[hard,] received, process exiting"
}

trap cleanup SIGINT SIGTERM

echo "Fri Apr 12 03:40:00 2019 OpenVPN 2.4.5 aarch64-openwrt-linux-gnu [SSL (OpenSSL)] [LZO] [LZ4] [EPOLL] [MH/PKTINFO] [AEAD]"
echo "Fri Apr 12 03:40:00 2019 library versions: OpenSSL 1.0.2r  26 Feb 2019, LZO 2.10"
echo "Fri Apr 12 03:40:00 2019 Outgoing Control Channel Authentication: Using 512 bit message hash 'SHA512' for HMAC authentication"
echo "Fri Apr 12 03:40:00 2019 Incoming Control Channel Authentication: Using 512 bit message hash 'SHA512' for HMAC authentication"
sleep 4
echo "Fri Apr 12 03:40:04 2019 TCP/UDP: Preserving recently used remote address: [AF_INET]1.2.3.4:1194"
echo "Fri Apr 12 03:40:04 2019 UDP link local: (not bound)"
echo "Fri Apr 12 03:40:04 2019 UDP link remote: [AF_INET]1.2.3.4:1194"
sleep 1
echo "Fri Apr 12 03:40:05 2019 WARNING: this configuration may cache passwords in memory -- use the auth-nocache option to prevent this"
echo "Fri Apr 12 03:40:05 2019 VERIFY OK: depth=1, C=CA, ST=ON, L=Toronto, O=VPN Limited, OU=Operations, CN=VPN Node CA"
echo "Fri Apr 12 03:40:05 2019 VERIFY KU OK"
echo "Fri Apr 12 03:40:05 2019 Validating certificate extended key usage"
echo "Fri Apr 12 03:40:05 2019 ++ Certificate has EKU (str) TLS Web Server Authentication, expects TLS Web Server Authentication"
echo "Fri Apr 12 03:40:05 2019 VERIFY EKU OK"
echo "Fri Apr 12 03:40:05 2019 VERIFY OK: depth=0, C=CA, ST=ON, O=VPN Limited, OU=Operations, CN=VPN Node Server 4096"
echo "Fri Apr 12 03:40:05 2019 Control Channel: TLSv1.2, cipher TLSv1/SSLv3 ECDHE-RSA-AES256-GCM-SHA384, 4096 bit RSA"
echo "Fri Apr 12 03:40:05 2019 [VPN Node Server 4096] Peer Connection Initiated with [AF_INET]1.2.3.4:1194"
sleep 1
echo "Fri Apr 12 03:40:06 2019 Outgoing Data Channel: Cipher 'AES-256-GCM' initialized with 256 bit key"
echo "Fri Apr 12 03:40:06 2019 Incoming Data Channel: Cipher 'AES-256-GCM' initialized with 256 bit key"
echo "Fri Apr 12 03:40:06 2019 TUN/TAP device tun0 opened"
echo "Fri Apr 12 03:40:06 2019 do_ifconfig, tt->did_ifconfig_ipv6_setup=0"
echo "Fri Apr 12 03:40:06 2019 /sbin/ifconfig tun0 10.20.30.40 netmask 255.255.254.0 mtu 1500 broadcast 10.20.30.255"
echo "Fri Apr 12 03:40:06 2019 Initialization Sequence Completed"

while true
do
    test $ABORT -eq 1 && break
    sleep 1
done
