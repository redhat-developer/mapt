#!/bin/sh
set -eu

# Detect root LV (e.g. /dev/mapper/rootvg-rootlv)
ROOT_LV=$(findmnt -n -o SOURCE /)

# Get VG name
VG=$(lvs --noheadings -o vg_name "$ROOT_LV" | xargs)

# Get PV device (e.g. /dev/sda4 or /dev/nvme0n1p4)
PV=$(pvs --noheadings -o pv_name --select vg_name="$VG" | xargs)

# Extract base disk and partition number safely
DISK_PATH="$PV"

# NVMe devices end with pN (nvme0n1p4)
case "$DISK_PATH" in
  *nvme*)
    DISK=$(echo "$DISK_PATH" | sed -E 's/p[0-9]+$//')
    PART=$(echo "$DISK_PATH" | sed -E 's/.*p([0-9]+)$/\1/')
    ;;
  *)
    # Standard SCSI (sda4 → sda + 4)
    DISK=$(echo "$DISK_PATH" | sed -E 's/[0-9]+$//')
    PART=$(echo "$DISK_PATH" | sed -E 's/.*([0-9]+)$/\1/')
    ;;
esac

# Expand partition
growpart "$DISK" "$PART"

# Resize PV
pvresize "$PV"

# Extend LV to full free space and resize filesystem
lvextend -r -l +100%FREE "$ROOT_LV"