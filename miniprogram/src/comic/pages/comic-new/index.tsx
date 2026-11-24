import { useState } from "react";
import { readComicActiveRun, clearComicActiveRun } from "../../services/comicRecovery";
import ManualComicPrivatePage from "./ManualComicPrivatePage";
import LegacyComicPrivatePage from "./LegacyComicPrivatePage";
export { LegacyComicPrivatePage };
/** Only previously submitted automatic jobs retain their original recovery path. */
export default function ComicEntry() {
    const [initialSource, setInitialSource] = useState("");
    const [legacy, setLegacy] = useState(() => { const run = readComicActiveRun(); return Boolean(run?.submitted && run.mode !== "manual"); });
    return legacy ? <LegacyComicPrivatePage onNew={source => { clearComicActiveRun(); setInitialSource(source || ""); setLegacy(false); }}/> : <ManualComicPrivatePage initialSource={initialSource}/>;
}
