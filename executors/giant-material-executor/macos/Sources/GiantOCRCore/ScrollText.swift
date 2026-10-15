import Foundation

public struct FrameLine {
    public let text: String
    public let y: Double
    public let height: Double

    public init(text: String, y: Double, height: Double) {
        self.text = text
        self.y = y
        self.height = height
    }
}

public struct ScrollMerge {
    public let text: String
    public let duplicate: Bool
}

private func isContent(_ character: Character) -> Bool {
    character.unicodeScalars.contains {
        CharacterSet.letters.contains($0) || CharacterSet.decimalDigits.contains($0)
    }
}

private func normalizedCharacters(_ value: String) -> [Character] {
    value.filter(isContent)
}

public func filterFrameLines(_ lines: [FrameLine], seconds: Int) -> String {
    lines.filter { line in
        let text = line.text.trimmingCharacters(in: .whitespacesAndNewlines)
        guard line.y >= 0.12, line.y + line.height <= (seconds == 0 ? 0.94 : 0.90),
              text.contains(where: isContent) else { return false }
        if ["热门小说", "点击下方链接", "免费获取全文"].contains(text) { return false }
        if ["本故事纯属虚构", "无不良引导", "危險动作请勿模仿", "危险动作请勿模仿", "危脸动作请勿模仿"].contains(where: text.hasPrefix) { return false }
        return true
    }.map(\.text).joined(separator: "\n")
}

public func cleanText(_ value: String) -> String {
    value.replacingOccurrences(of: "\r\n", with: "\n")
        .replacingOccurrences(of: "\r", with: "\n")
        .components(separatedBy: "\n")
        .map { $0.trimmingCharacters(in: .whitespacesAndNewlines) }
        .filter { $0.isEmpty || $0.contains(where: isContent) }
        .joined(separator: "\n")
        .trimmingCharacters(in: .whitespacesAndNewlines)
}

private func overlapLength(previous: String, current: String) -> Int {
    let left = Array(normalizedCharacters(previous).suffix(600))
    let right = Array(normalizedCharacters(current).prefix(600))
    let width = right.count
    guard left.count >= 20, width >= 20 else { return 0 }
    var row = Array(0...width)
    for source in left {
        var next = Array(repeating: 0, count: width + 1)
        for column in 1...width {
            next[column] = min(row[column] + 1,
                               next[column - 1] + 1,
                               row[column - 1] + (source == right[column - 1] ? 0 : 1))
        }
        row = next
    }
    var best = 0
    var score = 19
    for column in 20...width where Double(row[column]) / Double(column) <= 0.12 {
        let candidate = column - 5 * row[column]
        if candidate > score {
            score = candidate
            best = column
        }
    }
    return best
}

public func appendScrollText(body: String, previous: String, current: String) -> ScrollMerge {
    let current = current.trimmingCharacters(in: .whitespacesAndNewlines)
    if current.isEmpty { return ScrollMerge(text: body, duplicate: true) }
    if body.isEmpty { return ScrollMerge(text: current, duplicate: false) }

    let normalizedCurrent = normalizedCharacters(current)
    let normalizedTail = normalizedCharacters(String(body.suffix(2500)))
    if !normalizedCurrent.isEmpty,
       normalizedTail.count >= normalizedCurrent.count,
       String(normalizedTail).contains(String(normalizedCurrent)) {
        return ScrollMerge(text: body, duplicate: true)
    }

    let matched = overlapLength(previous: previous, current: current)
    guard matched > 0 else { return ScrollMerge(text: body + "\n\n" + current, duplicate: false) }
    var contentSeen = 0
    var boundary = current.startIndex
    for index in current.indices {
        if isContent(current[index]) {
            contentSeen += 1
            if contentSeen == matched {
                boundary = current.index(after: index)
                break
            }
        }
    }
    let extra = String(current[boundary...])
    if !extra.contains(where: isContent) { return ScrollMerge(text: body, duplicate: true) }
    guard let finalContent = body.indices.last(where: { isContent(body[$0]) }) else {
        return ScrollMerge(text: body + extra, duplicate: false)
    }
    return ScrollMerge(text: String(body[...finalContent]) + extra, duplicate: false)
}
