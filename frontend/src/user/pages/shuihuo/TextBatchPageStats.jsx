export default function TextBatchPageStats({books = [], summaries = {}}) {
  const counts = books.reduce((sum, book) => ({
    ready:sum.ready + Number(summaries[book.id]?.textImageReadyCount || 0),
    total:sum.total + Number(summaries[book.id]?.textImageTotalCount || 0)
  }), {ready:0,total:0});
  return <span className="is-uploaded"><b>{counts.ready}/{counts.total}</b> 文字图片</span>;
}
