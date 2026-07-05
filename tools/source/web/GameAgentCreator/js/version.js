/* ============= Version Check ============= */
const CREATOR_MIN_COMPATIBLE = "v0.4.0";

let engineVersionInfo = null;

async function checkEngineVersion() {
  try {
    const info = await api("GET", "/api/v1/version");
    engineVersionInfo = info;
    // Parse versions
    const cv = parseSemVer(CREATOR_MIN_COMPATIBLE);
    const ev = parseSemVer(info.version);
    if (!cv || !ev) return; // can't parse, skip check
    // Check: major and minor must match, patch >= min
    if (cv.major !== ev.major || cv.minor !== ev.minor) {
      toast("版本不兼容: Creator需要" + CREATOR_MIN_COMPATIBLE + "，Engine运行" + info.version, "error");
      console.warn("Version mismatch: Creator requires", CREATOR_MIN_COMPATIBLE, "Engine is", info.version);
    } else if (ev.patch < cv.patch) {
      toast("Engine版本过旧: 需要" + CREATOR_MIN_COMPATIBLE + "，当前" + info.version, "error");
    }
    return info;
  } catch(e) {
    console.warn("Cannot check engine version:", e);
    return null;
  }
}

function parseSemVer(v) {
  const s = v.replace(/^v/i, "");
  const parts = s.split(".");
  if (parts.length !== 3) return null;
  return { major: parseInt(parts[0]), minor: parseInt(parts[1]), patch: parseInt(parts[2]) };
}
