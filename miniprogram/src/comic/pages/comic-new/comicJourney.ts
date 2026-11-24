/** This is optional example story content, never a system prompt or an automatically submitted task. */
export const COMIC_EXAMPLE_SOURCE = "清晨，林夏推开旧书店的门。阳光落在柜台上的信封上，店主说，这封信已经等了她二十年。她拆开信封，里面是一张旧照片，照片上的女孩与她长得一模一样。\n\n照片背面写着：日落前，到钟楼来。林夏赶到钟楼，发现墙上挂着母亲年轻时的画。她循着画中的街道，走进一间正在修复老照片的工作室。\n\n一位老人取出最后一张照片，笑着告诉她，母亲曾在这里留下一个约定：等女儿长大，就把这条街的故事讲给她听。林夏终于明白，这封迟到的信，是母亲为她准备的一次重逢。";

type PageReadiness = {primaryImageId: string; images: Array<{id: string}>; status: string};
export function comicPageReadiness(pages: PageReadiness[]) {
  let ready = 0; let failed = 0; let firstFailedIndex = -1;
  pages.forEach((page, index) => {
    if (page.primaryImageId && page.images.some(image => image.id === page.primaryImageId)) ready++;
    else if (/fail|error/i.test(page.status)) { failed++; if (firstFailedIndex < 0) firstFailedIndex = index; }
  });
  return {total: pages.length, ready, failed, waiting: pages.length - ready - failed, firstFailedIndex};
}
