import Foundation

@main
struct ScrollTextSelfTest {
    static func main() throws {
        let first = "第一句很长的滚屏小说正文，已经超过二十个字符。\n第二句也在屏幕中。"
        let next = first + "\n第三句是新出现的。"
        let merged = appendScrollText(body: first, previous: first, current: next)
        precondition(merged.text == next && !merged.duplicate, "overlap must append only new text")

        let repeated = appendScrollText(body: first, previous: first, current: first)
        precondition(repeated.text == first && repeated.duplicate, "repeated frame must be skipped")

        let distinct = appendScrollText(body: "第一段。", previous: "第一段。", current: "另一段完全不同的内容。")
        precondition(distinct.text == "第一段。\n\n另一段完全不同的内容。", "distinct text must be preserved")
        let noisyPrevious = "这是第一段完整的滚屏正文，包含足够多的字用于匹配。第二行继续。"
        let noisyCurrent = "这是第一段完整的滚屏正文，包含足够多的字用于匹配。第二行继读。新的一句。"
        let noisyMerged = appendScrollText(body: noisyPrevious, previous: noisyPrevious, current: noisyCurrent)
        precondition(noisyMerged.text == noisyPrevious + "新的一句。", "one OCR typo must not duplicate an overlapping frame")
        let lines = [
            FrameLine(text: "热门小说", y: 0.86, height: 0.04),
            FrameLine(text: "第一行小说正文。", y: 0.70, height: 0.04),
            FrameLine(text: "第二行小说正文。", y: 0.22, height: 0.04),
            FrameLine(text: "点击下方链接", y: 0.05, height: 0.04)
        ]
        precondition(filterFrameLines(lines, seconds: 1) == "第一行小说正文。\n第二行小说正文。", "page chrome must not enter novel body")
        precondition(cleanText("第一句。\n。\n第二句。") == "第一句。\n第二句。", "punctuation-only line must be removed")

        do {
            _ = try validateExtractRequest(jobID: "job", videoURL: "https://evil.example/a.mp4", duration: 12)
            preconditionFailure("other hosts must be rejected")
        } catch {}
        do {
            _ = try validateExtractRequest(jobID: "job", videoURL: "https://material.hnqingyuwen.top/a.mp4", duration: 1801)
            preconditionFailure("long videos must be rejected")
        } catch {}
        _ = try validateExtractRequest(jobID: "job", videoURL: "https://material.hnqingyuwen.top/a.mp4", duration: 12)
        _ = try validateExtractRequest(jobID: "job", videoURL: "https://mlzr-material.hnqingyuwen.top/a.mp4", duration: 12)
        _ = try validateExtractRequest(jobID: "job", videoURL: "https://ml-material.hnqingyuwen.top/a.mp4", duration: 12)
        print("ScrollTextSelfTest passed")
    }
}
