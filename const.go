package main

import "regexp"

var shortPortNameRegex = regexp.MustCompile(`\d+\/.+`)

// 802.11 reason codes (IEEE Std 802.11-2012, Table 8-36)
var WiFiReasonCodes = map[string]string{
	"0":  "Reserved",
	"1":  "Unspecified reason",
	"2":  "Previous authentication no longer valid",
	"3":  "Deauthenticated because sending STA is leaving (or has left) IBSS or ESS",
	"4":  "Disassociated due to inactivity",
	"5":  "Disassociated because AP is unable to handle all currently associated STAs",
	"6":  "Class 2 frame received from nonauthenticated STA",
	"7":  "Class 3 frame received from nonassociated STA",
	"8":  "Disassociated because sending STA is leaving (or has left) BSS",
	"9":  "STA requesting (re)association is not authenticated with responding STA",
	"10": "Disassociated because the information in the Power Capability element is unacceptable",
	"11": "Disassociated because the information in the Supported Channels element is unacceptable",
	"12": "Disassociated due to BSS Transition Management",
	"13": "Invalid element, i.e., an element defined in this standard for which the content does not meet the specifications",
	"14": "Message integrity code (MIC) failure",
	"15": "4-Way Handshake timeout",
	"16": "Group Key Handshake timeout",
	"17": "element in 4-Way Handshake different from (Re)Association Request/Probe Response/Beacon frame",
	"18": "Invalid group cipher",
	"19": "Invalid pairwise cipher",
	"20": "Invalid AKMP",
	"21": "Unsupported RSNE version",
	"22": "Invalid RSNE capabilities",
	"23": "IEEE 802.1X authentication failed",
	"24": "Cipher suite rejected because of the security policy",
	"25": "TDLS direct-link teardown due to TDLS peer STA unreachable via the TDLS direct link",
	"26": "TDLS direct-link teardown for unspecified reason",
	"27": "Disassociated because session terminated by SSP request",
	"28": "Disassociated because of lack of SSP roaming agreement",
	"29": "Requested service rejected because of SSP cipher suite or AKM requirement",
	"30": "Requested service not authorized in this location",
	"31": "TS deleted because QoS AP lacks sufficient bandwidth for this QoS STA due to a change in BSS service characteristics or operational mode (e.g., an HT BSS change from 40 MHz channel to 20 MHz channel)",
	"32": "Disassociated for unspecified, QoS-related reason",
	"33": "Disassociated because QoS AP lacks sufficient bandwidth for this QoS STA",
	"34": "Disassociated because excessive number of frames need to be acknowledged, but are not acknowledged due to AP transmissions and/or poor channel conditions",
	"35": "Disassociated because STA is transmitting outside the limits of its TXOPs",
	"36": "Requested from peer STA as the STA is leaving the BSS (or resetting)",
	"37": "Requested from peer STA as it does not want to use the mechanism",
	"38": "Requested from peer STA as the STA received frames using the mechanism for which a setup is required",
	"39": "Requested from peer STA due to timeout",
	"45": "Peer STA does not support the requested cipher suite",
	"46": "In a DLS Teardown frame: The teardown was initiated by the DLS peer. In a Disassociation frame: Disassociated because authorized access limit reached",
	"47": "In a DLS Teardown frame: The teardown was initiated by the AP. In a Disassociation frame: Disassociated due to external service requirements",
	"48": "Invalid FT Action frame count",
	"49": "Invalid pairwise master key identifier (PMKI)",
	"50": "Invalid MDE",
	"51": "Invalid FTE",
	"52": "SME cancels the mesh peering instance with the reason other than reaching the maximum number of peer mesh STAs",
	"53": "The mesh STA has reached the supported maximum number of peer mesh STAs",
	"54": "The received information violates the Mesh Configuration policy configured in the mesh STA profile",
	"55": "The mesh STA has received a Mesh Peering Close message requesting to close the mesh peering",
	"56": "The mesh STA has resent dot11MeshMaxRetries Mesh Peering Open messages, without receiving a Mesh Peering Confirm message",
	"57": "The confirmTimer for the mesh peering instance times out",
	"58": "The mesh STA fails to unwrap the GTK or the values in the wrapped contents do not match",
	"59": "The mesh STA receives inconsistent information about the mesh parameters between Mesh Peering Management frames",
	"60": "The mesh STA fails the authenticated mesh peering exchange because due to failure in selecting either the pairwise ciphersuite or group ciphersuite",
	"61": "The mesh STA does not have proxy information for this external destination",
	"62": "The mesh STA does not have forwarding information for this destination",
	"63": "The mesh STA determines that the link to the next hop of an active path in its forwarding information is no longer usable",
	"64": "The Deauthentication frame was sent because the MAC address of the STA already exists in the mesh BSS",
	"65": "The mesh STA performs channel switch to meet regulatory requirements",
	"66": "The mesh STA performs channel switch with unspecified reason",
}
