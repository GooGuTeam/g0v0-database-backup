#!/usr/bin/env bash

set -euo pipefail

usage() {
  cat <<'EOF'
用法:
  migrate_rclone_legacy_backups.sh <remote> [source_base]

参数:
  remote       rclone 远端名称，例如 onedrive:
  source_base  旧备份所在目录，默认为 backup

环境变量:
  RCLONE_CONFIG   可选，自定义 rclone.conf 路径
  DRY_RUN=1       仅打印将要执行的 moveto，不真正移动

说明:
  该脚本会扫描 remote/source_base 下的旧平铺备份目录:
    db_YYYYMMDD_HHMM
    db_YYYYMMDD_HHMM_inc
  并将其迁移为:
    source_base/YYMM/DD/db_YYYYMMDD_HHMM[_inc]
EOF
}

if [[ $# -lt 1 || $# -gt 2 ]]; then
  usage
  exit 1
fi

REMOTE="$1"
SOURCE_BASE="${2:-backup}"
SOURCE_BASE="${SOURCE_BASE#/}"
SOURCE_BASE="${SOURCE_BASE%/}"

if [[ -z "$REMOTE" || -z "$SOURCE_BASE" ]]; then
  usage
  exit 1
fi

RCLONE_ARGS=()
if [[ -n "${RCLONE_CONFIG:-}" ]]; then
  RCLONE_ARGS+=(--config "$RCLONE_CONFIG")
fi

mapfile -t entries < <(rclone "${RCLONE_ARGS[@]}" lsf "${REMOTE}${SOURCE_BASE}" --dirs-only)

for raw_entry in "${entries[@]}"; do
  entry="${raw_entry%/}"
  if [[ ! "$entry" =~ ^db_([0-9]{4})([0-9]{2})([0-9]{2})_[0-9]{4}(_inc)?$ ]]; then
    echo "跳过非旧格式目录: ${entry}"
    continue
  fi

  yy_mm="${BASH_REMATCH[1]:2:2}${BASH_REMATCH[2]}"
  dd="${BASH_REMATCH[3]}"

  source_path="${REMOTE}${SOURCE_BASE}/${entry}"
  target_path="${REMOTE}${SOURCE_BASE}/${yy_mm}/${dd}/${entry}"

  if [[ "${DRY_RUN:-0}" == "1" ]]; then
    echo "DRY RUN: rclone ${RCLONE_ARGS[*]} moveto ${source_path} ${target_path}"
    continue
  fi

  echo "移动 ${source_path} -> ${target_path}"
  rclone "${RCLONE_ARGS[@]}" moveto "${source_path}" "${target_path}"
done
