/* ============= Version Check ============= */
const CREATOR_MIN_COMPATIBLE = "v0.4.5";

let engineVersionInfo = null;

async function checkEngineVersion() {
  try {
    const info = await api("GET", "/api/v1/version");
    engineVersionInfo = info;
    const cv = parseSemVer(CREATOR_MIN_COMPATIBLE);
    const ev = parseSemVer(info.version);
    if (!cv || !ev) return;
    if (cv.major !== ev.major || cv.minor !== ev.minor) {
      toast("Creator requires " + CREATOR_MIN_COMPATIBLE + ", engine is running " + info.version, "error");
      console.warn("Version mismatch: Creator requires", CREATOR_MIN_COMPATIBLE, "Engine is", info.version);
    } else if (ev.patch < cv.patch) {
      toast("Engine version is too old: requires " + CREATOR_MIN_COMPATIBLE + ", current " + info.version, "error");
    }
    return info;
  } catch (e) {
    console.warn("Cannot check engine version:", e);
    return null;
  }
}

function parseSemVer(v) {
  const s = v.replace(/^v/i, "");
  const parts = s.split(".");
  if (parts.length !== 3) return null;
  return { major: parseInt(parts[0], 10), minor: parseInt(parts[1], 10), patch: parseInt(parts[2], 10) };
}
