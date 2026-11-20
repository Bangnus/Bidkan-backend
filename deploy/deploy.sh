#!/usr/bin/env bash
# ขึ้นเวอร์ชันใหม่ของ Bidkan-backend บน Droplet: bash /opt/bidkan/deploy.sh <commit 40 ตัว>
# GitHub Actions เรียกผ่าน SSH ด้วย key ที่ถูกบังคับให้รันได้แค่สคริปต์นี้ (command= ใน /root/.ssh/authorized_keys — DEPLOYMENT.md ข้อ 3.4)
# คำสั่งที่ส่งมาทาง SSH อยู่ใน SSH_ORIGINAL_COMMAND จึงรับได้แค่เลข commit: key หลุดก็สั่งอย่างอื่นไม่ได้
# ไฟล์ตั้งค่า (compose, caddy, mosquitto.conf, schema, สคริปต์นี้) มากับ image bidkan-deploy-files บน GHCR เลข commit เดียวกัน
set -euo pipefail

# ทั้งสคริปต์อยู่ในฟังก์ชัน: รอบนี้อาจเขียนไฟล์นี้ทับระหว่างรัน bash จึงต้องอ่านทั้งไฟล์ให้จบก่อน
main() {
  local sha
  read -r sha _ <<<"${SSH_ORIGINAL_COMMAND:-$*}"
  if [[ ! "${sha:-}" =~ ^[0-9a-f]{40}$ ]]; then
    echo "usage: deploy.sh <full 40-character commit hash>" >&2
    exit 2
  fi

  cd "$(dirname "$(readlink -f "$0")")"
  local compose=(docker compose -p bidkan -f docker-compose.prod.yml --env-file .env.prod)

  # อ่านแค่บรรทัดนี้จาก .env.prod (ไม่ source ไฟล์ที่มีค่าลับ)
  local owner
  owner="$(sed -n 's/^GHCR_OWNER=//p' .env.prod | tail -n 1 | tr -d "\"' \r")"
  if [[ ! "$owner" =~ ^[a-z0-9][a-z0-9-]*$ ]]; then
    echo "GHCR_OWNER in .env.prod must be the lowercase GitHub owner" >&2
    exit 2
  fi

  # ไม่มีไฟล์ = docker สร้างเป็นโฟลเดอร์แทน แล้ว API ใช้ Firebase ไม่ได้ (OTP/แจ้งเตือนไม่ออก) จึงหยุดตั้งแต่ตรงนี้
  if [[ ! -f /opt/bidkan/firebase-service-account.json ]]; then
    echo "/opt/bidkan/firebase-service-account.json is missing (DEPLOYMENT.md step 3.3)" >&2
    exit 2
  fi

  exec 9>/var/lock/bidkan-deploy.lock
  flock -w 1800 9 || { echo "another deploy is still running" >&2; exit 1; }

  # ---- ไฟล์ตั้งค่าจาก GHCR: ดึงครบและตรวจก่อน ค่อยวางทับของเดิม ----
  local files_image="ghcr.io/$owner/bidkan-deploy-files:$sha" tmp cid f
  local files=(docker-compose.prod.yml bidkan.caddy mosquitto.conf deploy.sh)
  echo "== pull $files_image"
  docker pull --quiet "$files_image" >/dev/null
  tmp="$(mktemp -d)"
  # image ไม่มีโปรแกรมอะไรเลย (FROM scratch) สร้าง container ไว้คัดลอกไฟล์ออกเฉย ๆ ไม่ได้รัน
  cid="$(docker create "$files_image" /none)"
  for f in "${files[@]}"; do
    if ! docker cp "$cid:/$f" "$tmp/$f" >/dev/null || [[ ! -s "$tmp/$f" ]]; then
      echo "$f is missing from $files_image" >&2
      docker rm "$cid" >/dev/null
      rm -rf "$tmp"
      exit 1
    fi
  done
  if ! docker cp "$cid:/schema" "$tmp/schema" >/dev/null; then
    echo "schema/ is missing from $files_image" >&2
    docker rm "$cid" >/dev/null
    rm -rf "$tmp"
    exit 1
  fi
  docker rm "$cid" >/dev/null
  # เจ้าของ root และเขียนได้แค่ root · mv = สลับทั้งไฟล์ bash ที่รันอยู่ยังอ่านไฟล์เดิม
  for f in "${files[@]}"; do
    local mode=0644
    [[ "$f" == *.sh ]] && mode=0755
    install -o root -g root -m "$mode" "$tmp/$f" ".$f.new"
    mv -f ".$f.new" "$f"
  done
  # schema ใช้แค่ตอนสร้างฐานข้อมูลครั้งแรก — วางทับทั้งโฟลเดอร์ได้เลย
  chown -R root:root "$tmp/schema"
  chmod -R a+rX,go-w "$tmp/schema"
  rm -rf schema
  mv "$tmp/schema" schema
  rm -rf "$tmp"
  docker rmi "$files_image" >/dev/null 2>&1 || true

  # ---- API ----
  export IMAGE_TAG="$sha"
  echo "== pull bidkan-backend:$sha"
  "${compose[@]}" pull --quiet bidkan-api
  echo "== start"
  "${compose[@]}" up -d --wait --wait-timeout 180
  docker image prune -f >/dev/null

  # ---- โดเมน API ใน Caddy กลางของเครื่อง (/opt/edge): ใส่ไฟล์ใหม่ ตรวจ config ทั้งเครื่อง ผ่านค่อย reload ----
  # ไม่ผ่าน = install-site.sh คืนไฟล์เดิมแล้ว exit 1 (ไฟล์ผิดที่ค้างไว้ทำให้ Caddy เปิดไม่ขึ้น ทุกเว็บบนเครื่องล่ม)
  /opt/edge/install-site.sh bidkan bidkan.caddy

  "${compose[@]}" ps
  echo "== done: bidkan-backend is at $sha"
}

main "$@"
exit
