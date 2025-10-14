docker run --platform linux/amd64 \
  --rm -it \
  -p3000:3000 \
  -e PORT=3000 \
  -e DAPLA_TEAM_API_URL=https://dapla-team-api.intern.test.ssb.no \
  -e DAPLA_CTRL_ADMIN_GROUPS=dapla-stat-developers,dapla-skyinfra-developers \
  -e DAPLA_CTRL_DOCUMENTATION_URL=https://manual.dapla.ssb.no/statistikkere/dapla-ctrl.html \
  dapla-ctrl:latest
