import {apiRequest} from '../../../services/api';
import {sessionRevision,onSessionInvalidated} from '../../../services/sessionBoundary';
import {saveComicSelectedBook,comicRunStorageKey,prepareComicRecovery} from '../../services/comicRecovery';
import Taro, { useRouter } from "@tarojs/taro";
import { Button, Image, ScrollView, Text, View } from "@tarojs/components";
import { useEffect, useRef, useState } from "react";
import BrandHeader from "../../../components/BrandHeader";
import { ThemedPage } from "../../../theme";
import ComicStickySpacer from "../../components/ComicStickySpacer";
import "../../styles/comic.scss";
import {
  CLIENT_REFERENCE_LIMIT_BYTES, DEFAULT_COMIC_ASPECT_RATIO, DEFAULT_COMIC_RUN_MODE,
  comicBooksWithoutSource, loadComicBooks, loadComicSetup, resolveComicStyleThumb, submitComicRun, uploadComicReference,
  type ComicBookOption, type ComicReferenceInfo, type ComicSetup, type ComicStylePreset,
} from "../../services/comicFlow";

type FormState = {
  batchIndex: number;
  bookIds: string[];
  styleIndex: number;
  aspectIndex: number;
  referencePath: string;
  reference: ComicReferenceInfo | null;
};

const ASPECT_CHOICES = [
  { value: "9:16", label: "竖版", note: "适合手机阅读" },
  { value: "1:1", label: "方形", note: "适合社交分享" },
  { value: "16:9", label: "横版", note: "适合宽幅画面" },
];

const INITIAL_FORM: FormState = {
  batchIndex: 0, bookIds: [], styleIndex: 0, aspectIndex: 0, referencePath: "", reference: null,
};

/** 记住上一次做漫画用的批次，下次进来直接用它，不再先问「选哪个批次」。 */
const LAST_BATCH_STORAGE_KEY = "comic:last-batch-id";

function readRememberedBatchId() {
  try {
    const value = Taro.getStorageSync(LAST_BATCH_STORAGE_KEY);
    return typeof value === "string" ? value : "";
  } catch {
    return "";
  }
}

function rememberBatch(batchId: string) {
  try { Taro.setStorageSync(LAST_BATCH_STORAGE_KEY, batchId); } catch { /* 存储失败不影响本次选择。 */ }
}

function Step({ index, title, hint, children }: { index: string; title: string; hint?: string; children?: any }) {
  return (
    <View className="comic-step">
      <View className="comic-step-head">
        <Text className="comic-step-index">{index}</Text>
        <Text className="comic-step-title">{title}</Text>
      </View>
      {hint ? <Text className="comic-step-hint">{hint}</Text> : null}
      {children ? <View className="comic-step-body">{children}</View> : null}
    </View>
  );
}

export default function ComicConfigPage() {
  const router = useRouter();
  // 从「开始做漫画」页新建成功后会把新作品带过来，直接定位到它，省掉再选一次。
  const requestedBatchId = String(router?.params?.batchId || "");
  const requestedBookId = String(router?.params?.bookId || "");
  const pickMode = String(router?.params?.mode || "") === "pick";
  const [setup, setSetup] = useState<ComicSetup | null>(null);
  const [books, setBooks] = useState<ComicBookOption[]>([]);
  const [form, setForm] = useState<FormState>(INITIAL_FORM);
  const [batchPickerOpen, setBatchPickerOpen] = useState(false);
  const [showMore, setShowMore] = useState(false);
  const [styleThumbs, setStyleThumbs] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(true);
  const [loadingBooks, setLoadingBooks] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");
  const version=useRef(0),accountRevision=useRef(sessionRevision());

  const loadBooks = async (batchId: string, batchIndex: number, preferredBookId = "") => {
    const current = version.current;
    setLoadingBooks(true);
    setBooks([]);
    setForm((value) => ({ ...value, batchIndex, bookIds: [], referencePath: "", reference: null }));
    rememberBatch(batchId);
    try {
      const next = await loadComicBooks(batchId);
      if (current !== version.current) return;
      setBooks(next);
      // 多本书不默认全选：一次全跑会直接产生多份生成消耗，默认只选中第一本；带 bookId 进来则直达那本。
      const target = preferredBookId && next.some((book) => book.id === preferredBookId) ? preferredBookId : next[0]?.id;
      setForm((value) => ({ ...value, bookIds: target ? [target] : [] }));
    } catch (cause: any) {
      if (current === version.current) setError(cause?.message || "没有读到这个批次里的书，请稍后再试");
    } finally {
      if (current === version.current) setLoadingBooks(false);
    }
  };

  // 缩略图先占位、再带鉴权下载到本地显示；下载失败就保持占位，不让它变成红色网络错误。
  const loadStyleThumbs = async (list: ComicStylePreset[]) => {
    const current = version.current;
    await Promise.all(list.map(async (style) => {
      if (!style.previewUrl) return;
      const path = await resolveComicStyleThumb(style.previewUrl);
      if (current === version.current && path) setStyleThumbs((value) => ({ ...value, [style.id]: path }));
    }));
  };

  const load=()=>{
    accountRevision.current=sessionRevision();setSetup(null);setBooks([]);
    const current = ++version.current;
    setLoading(true);
    setError("");
    setStyleThumbs({});
    void loadComicSetup()
      .then(async (result) => {
        if (current !== version.current) return;
        setSetup(result);
        void loadStyleThumbs(result.styles);
        // 选批次顺序：URL 指定 → 本地记住的最近批次 → 列表第一项。失效的 id 自动跳过。
        const requestedIndex = requestedBatchId ? result.batches.findIndex((batch) => batch.id === requestedBatchId) : -1;
        const rememberedId = readRememberedBatchId();
        const rememberedIndex = rememberedId ? result.batches.findIndex((batch) => batch.id === rememberedId) : -1;
        const index = requestedIndex >= 0 ? requestedIndex : rememberedIndex >= 0 ? rememberedIndex : 0;
        if (result.batches[index]) await loadBooks(result.batches[index].id, index, requestedBookId);
      })
      .catch((cause) => { if (current === version.current) { setSetup(null); setError(cause?.message || "暂时读不到作品，请稍后再试"); } })
      .finally(() => { if (current === version.current) setLoading(false); });
  };

  useEffect(()=>{load();const off=onSessionInvalidated(()=>{version.current++;accountRevision.current=-1;setSetup(null);setBooks([]);setForm(INITIAL_FORM);setLoading(false);setLoadingBooks(false);setError('账号已切换，请重新读取当前账号作品');});return()=>{version.current++;off();};},[]);

  const batches = setup?.batches || [];
  const styles = setup?.styles || [];
  const presets = setup?.presets || [];
  const currentBatch = batches[form.batchIndex] || batches[0];
  const currentBookCount = currentBatch?.bookCount || 0;
  const missingSourceBooks = comicBooksWithoutSource(books, form.bookIds);
  const canSubmit = Boolean(batches.length && books.length && styles.length && presets.length && !missingSourceBooks.length);

  const fillSource = (book: ComicBookOption) => {
    const batch = batches[form.batchIndex];
    const query = `batchId=${encodeURIComponent(batch?.id || "")}&bookId=${encodeURIComponent(book.id)}`
      + `&bookTitle=${encodeURIComponent(book.title)}&revision=${book.revision}`;
    void Taro.navigateTo({ url: `/comic/pages/comic-new/index?${query}` });
  };

  const selectBatch = (index: number) => {
    setBatchPickerOpen(false);
    if (index === form.batchIndex) return;
    void loadBooks(batches[index].id, index);
  };

  /** 小程序里没有「往已有批次加书」的入口，批量工厂的导入会新建批次，所以只跳首页，不假装能加书。 */
  const goBatchManage = () => {
    void Taro.navigateTo({ url: "/pages/batch/index" });
  };

  const toggleBook = (id: string) => setForm((value) => ({
    ...value,
    bookIds: value.bookIds.includes(id) ? value.bookIds.filter((item) => item !== id) : [...value.bookIds, id],
  }));

  const selectAllBooks = () => setForm((value) => ({ ...value, bookIds: books.map((book) => book.id) }));

  const chooseReference = async () => {
    setError("");
    try {
      const result = await Taro.chooseMedia({ count: 1, mediaType: ["image"], sizeType: ["compressed"] });
      const file = result.tempFiles?.[0];
      if (!file?.tempFilePath) return;
      if (typeof file.size === "number" && file.size > CLIENT_REFERENCE_LIMIT_BYTES) {
        setError("参考图太大了，请换一张小于 3 MB 的图片");
        return;
      }
      setForm((value) => ({ ...value, referencePath: file.tempFilePath, reference: null }));
    } catch {
      // 用户取消选择时不提示错误。
    }
  };

  const pickExistingBook=async(book:ComicBookOption)=>{try{if(accountRevision.current!==sessionRevision())throw Error('账号已切换，请重新读取已有作品');
    if (!book.sourceText.trim()) { setError(`《${book.title}》还没有正文，暂时不能选择`); return; }
    const revision=sessionRevision();if(!comicRunStorageKey()){const me:any=await apiRequest('/api/member/me');if(!me?.member?.username)throw Error('账号尚未确认，请重新打开漫画生成');await prepareComicRecovery(apiRequest,me.member.username);}if(revision!==sessionRevision())throw Error('账号已切换，请重新读取已有作品');
    saveComicSelectedBook({
      batchId: currentBatch?.id || "",
      bookId: book.id,
      title: book.title,
      sourceText: book.sourceText,
    });
    void Taro.navigateBack({delta:1});
    }catch(c:any){setError(c.message||'正文选择失败，请重新读取');}
  };

  const submit = async () => {
    if (submitting || !setup) return;
    const batch = batches[form.batchIndex];
    const preset = presets[0];
    const style = styles[form.styleIndex];
    if (!batch) { setError("请先选择一个批次"); return; }
    if (!preset) { setError("当前账号还没有生成权限，请联系管理员"); return; }
    if (!style) { setError("请选择一个画风"); return; }
    if (!form.bookIds.length) { setError("请至少选择一本书"); return; }
    // 后端分页只认正文，缺正文的书会在异步阶段被跳过，任务看着像"不动"，必须在这里拦住。
    const missing = comicBooksWithoutSource(books, form.bookIds);
    if (missing.length) { setError(`《${missing[0].title}》还没有正文，先补正文再生成`); return; }
    setSubmitting(true);
    setError("");
    try {
      const reference = form.reference || (form.referencePath ? await uploadComicReference(batch.id, form.referencePath) : null);
      await submitComicRun({
        batchId: batch.id,
        bookIds: form.bookIds,
        presetId: preset.id,
        runMode: DEFAULT_COMIC_RUN_MODE,
        aspectRatio: ASPECT_CHOICES[form.aspectIndex]?.value || DEFAULT_COMIC_ASPECT_RATIO,
        stylePresetId: style.id,
        reference,
      });
      const query = `batchId=${encodeURIComponent(batch.id)}&bookIds=${encodeURIComponent(form.bookIds.join(","))}`;
      void Taro.navigateTo({ url: `/comic/pages/comic-task/index?${query}` });
    } catch (cause: any) {
      setError(cause?.message || "没能开始生成，请稍后再试");
    } finally {
      setSubmitting(false);
    }
  };

  if (pickMode) return (
    <ThemedPage className="mp-page comic-page comic-book-picker-page">
      <BrandHeader eyebrow="漫画" title="选择已有书籍" showBack />
      <Text className="mp-title">选择已有书籍</Text>
      <Text className="mp-subtitle">选中后会把正文带回创作页，不会修改原书。</Text>
      {loading ? <View className="mp-card"><Text className="mp-row-title">正在读取书籍…</Text></View> : null}
      {!loading && setup ? <>
        <View className="comic-batch-summary">
          <Text className="comic-batch-line">{currentBatch?.title || "未命名作品"}</Text>
          {batches.length > 1 ? <Text className="comic-choice-mark" onClick={() => setBatchPickerOpen(!batchPickerOpen)}>{batchPickerOpen ? "收起" : "切换作品"}</Text> : null}
        </View>
        {batchPickerOpen ? <View className="comic-picker-list">{batches.map((batch, index) => <View key={batch.id} className={`comic-choice ${index === form.batchIndex ? "is-active" : ""}`} onClick={() => selectBatch(index)}><Text className="comic-choice-title">{batch.title}</Text><Text className="comic-choice-mark">{index === form.batchIndex ? "当前" : "切换"}</Text></View>)}</View> : null}
        <View className="comic-picker-list">
          {loadingBooks ? <Text className="comic-hint">正在读取正文…</Text> : books.map((book) => <View key={book.id} className={`comic-choice ${book.hasSource ? "" : "is-disabled"}`} onClick={() => pickExistingBook(book)}>
            <View><Text className="comic-choice-title">{book.title}</Text><Text className="comic-choice-note">{book.hasSource ? `${Array.from(book.sourceText.replace(/\s+/g, "")).length} 字` : "没有正文"}</Text></View>
            <Text className="comic-choice-mark">{book.hasSource ? "选择" : "不可选"}</Text>
          </View>)}
        </View>
        {error ? <Text className="mp-error">{error}</Text> : null}
      </> : null}
    </ThemedPage>
  );

  return (
    <ThemedPage className="mp-page comic-page">
      <BrandHeader eyebrow="漫画" title="做漫画" showBack />
      <Text className="mp-title">把小说做成漫画</Text>
      <Text className="mp-subtitle">批次已经帮你选好，确认书名、画风，点一下就开始。</Text>
      {loading ? <View className="mp-card"><Text className="mp-row-title">正在准备…</Text><Text className="mp-row-description">马上就好。</Text></View> : null}
      {!loading && error && !setup ? <View className="mp-card"><Text className="mp-row-title">暂时读不到批次</Text><Text className="mp-error">{error}</Text><Button className="mp-primary" onClick={load}>再试一次</Button></View> : null}
      {!loading && setup ? <>
        {batches.length ? (
          <View className="comic-batch-summary">
            <Text className="comic-batch-line">当前批次 · {currentBatch?.title || "未命名批次"}（共 {currentBookCount} 本书）</Text>
            {batches.length > 1 ? (
              <Text className="comic-choice-mark" onClick={() => setBatchPickerOpen((value) => !value)}>{batchPickerOpen ? "收起" : "切换"}</Text>
            ) : null}
          </View>
        ) : (
          <View className="comic-step">
            <Text className="comic-step-hint">还没有批次。粘贴一段小说正文就能新建一个。</Text>
            <Button className="mp-primary" onClick={() => { void Taro.navigateTo({ url: "/comic/pages/comic-new/index" }); }}>新建批次（粘贴正文开始）</Button>
            <Button className="mp-secondary" onClick={load}>从已有批次继续</Button>
            <Text className="comic-step-hint">在别处建过批次的话，点「从已有批次继续」重新读取一次。</Text>
          </View>
        )}

        {batchPickerOpen ? (
          <View className="comic-step">
            <Text className="comic-step-hint">切换到别的批次，这次会打到所选的批次里。</Text>
            {batches.map((batch, index) => (
              <View key={batch.id} className={`comic-choice ${index === form.batchIndex ? "is-active" : ""}`} onClick={() => selectBatch(index)}>
                <View>
                  <Text className="comic-choice-title">{batch.title}</Text>
                  <Text className="comic-choice-note">共 {batch.bookCount || 0} 本书</Text>
                </View>
                <Text className="comic-choice-mark">{index === form.batchIndex ? "当前" : "切换"}</Text>
              </View>
            ))}
          </View>
        ) : null}

        <Step index="1" title="选书" hint="一个批次里可以放多本书，都可以一起出图。">
          <View>
            {loadingBooks ? <Text className="comic-step-hint">正在读取这个批次里的书…</Text> : books.length > 1 ? <>
              <View className="comic-row-between">
                <Text className="comic-step-hint">已选 {form.bookIds.length} / {books.length} 本书</Text>
                <Text className="comic-choice-mark" onClick={selectAllBooks}>全选</Text>
              </View>
              {books.map((book) => (
                <View key={book.id} className={`comic-choice ${form.bookIds.includes(book.id) ? "is-active" : ""}`} onClick={() => toggleBook(book.id)}>
                  <Text className="comic-choice-title">{book.title}</Text>
                  <Text className="comic-choice-mark">{form.bookIds.includes(book.id) ? "已选" : "未选"}</Text>
                </View>
              ))}
            </> : books.length ? <Text className="comic-step-hint">这个批次只有 1 本书，已自动选中。</Text>
              : !batches.length ? null : <>
                <Text className="comic-step-hint">这个批次里还没有书。先去批量工厂处理，再回来。</Text>
                <Button className="mp-secondary" onClick={goBatchManage}>去批量工厂</Button>
              </>}
          </View>

          {missingSourceBooks[0] ? (
            <View className="comic-choice">
              <View style={{ flex: 1, minWidth: 0 }}>
                <Text className="comic-choice-title">《{missingSourceBooks[0].title}》还没有正文</Text>
                <Text className="comic-choice-note">
                  {missingSourceBooks.length > 1 ? `还有 ${missingSourceBooks.length - 1} 本也没有正文。` : ""}先补上正文，才能开始做漫画。
                </Text>
              </View>
              <Button className="mp-secondary" onClick={() => fillSource(missingSourceBooks[0])}>补正文</Button>
            </View>
          ) : null}
        </Step>

        <Step index="2" title="选画风" hint="画风决定整本的画面感觉。">
          {styles.length ? <ScrollView scrollX className="comic-style-row">
            {styles.map((style, index) => {
              const active = index === form.styleIndex;
              return (
                <View key={style.id} className={`comic-style-card ${active ? "is-active" : ""}`} onClick={() => setForm((value) => ({ ...value, styleIndex: index }))}>
                  {styleThumbs[style.id]
                    ? <Image src={styleThumbs[style.id]} mode="aspectFill" className="comic-style-thumb" />
                    : <View className="comic-style-fallback">{style.name.slice(0, 1)}</View>}
                  <Text className="comic-style-name">{style.name}</Text>
                  <Text className="comic-style-state">{active ? "已选中" : "点击选择"}</Text>
                </View>
              );
            })}
          </ScrollView> : <>
            <Text className="comic-step-hint">还没有可用的画风。先建一个画风，再回来做漫画。</Text>
            <Button className="mp-secondary" onClick={load}>重新读取</Button>
          </>}
        </Step>

        <Step index="3" title="加一张参考图（可选）" hint="如果有喜欢的画面感觉，可以传一张。不传也能生成。">
          <View>
            {form.referencePath ? <Image src={form.referencePath} lazyLoad mode="aspectFill" className="comic-reference" /> : null}
            <View className="comic-sticky-row">
              <Button className="mp-secondary" onClick={() => { void chooseReference(); }}>{form.referencePath ? "换一张" : "选择图片"}</Button>
              {form.referencePath ? <Button className="mp-secondary" onClick={() => setForm((value) => ({ ...value, referencePath: "", reference: null }))}>不用了</Button> : null}
            </View>
            <Text className="comic-step-hint">图片会保存到本地再上传，大小请控制在 3 MB 以内。</Text>
          </View>
        </Step>

        <View className="comic-step">
          <View className="comic-choice" onClick={() => setShowMore((value) => !value)}>
            <Text className="comic-choice-title">更多设置</Text>
            <Text className="comic-choice-mark">{showMore ? "收起" : "画幅 竖版"}</Text>
          </View>
          {showMore ? ASPECT_CHOICES.map((choice, index) => (
            <View key={choice.value} className={`comic-choice ${index === form.aspectIndex ? "is-active" : ""}`} onClick={() => setForm((value) => ({ ...value, aspectIndex: index }))}>
              <View>
                <Text className="comic-choice-title">{choice.label}</Text>
                <Text className="comic-choice-note">{choice.note}</Text>
              </View>
              <Text className="comic-choice-mark">{index === form.aspectIndex ? "已选中" : "切换"}</Text>
            </View>
          )) : null}
        </View>

        {error ? <Text className="mp-error">{error}</Text> : null}
        {!canSubmit ? <Text className="comic-step-hint">{missingSourceBooks.length ? "有书还没有正文，补上后才能开始。" : !styles.length ? "还没有可用的画风，先建一个画风再回来。" : !presets.length ? "当前账号还没有生成权限，请联系管理员开通。" : ""}</Text> : null}
      </> : null}

      <ComicStickySpacer />
      <View className="comic-sticky">
        {form.bookIds.length ? <Text className="comic-step-hint">将生成 {form.bookIds.length} 本书</Text> : null}
        <Button className="mp-primary" loading={submitting} disabled={loading || submitting || !canSubmit} onClick={() => { void submit(); }}>
          {submitting ? "正在开始…" : "开始生成"}
        </Button>
      </View>
    </ThemedPage>
  );
}
