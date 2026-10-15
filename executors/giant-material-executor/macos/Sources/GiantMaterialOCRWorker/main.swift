import Foundation
import AVFoundation
import Vision
import GiantOCRCore

private final class ResidentWorker: @unchecked Sendable {
    private let stateLock = NSLock()
    private let outputLock = NSLock()
    private let queue = DispatchQueue(label: "giant-material-vision-worker")
    private var activeJob: String?
    private var cancelled = false

    func emit(_ value: [String: Any]) {
        outputLock.lock()
        defer { outputLock.unlock() }
        guard let data = try? JSONSerialization.data(withJSONObject: value),
              var line = String(data: data, encoding: .utf8) else { return }
        line += "\n"
        FileHandle.standardOutput.write(Data(line.utf8))
    }

    func handle(_ raw: String) -> Bool {
        guard let data = raw.data(using: .utf8),
              let command = (try? JSONSerialization.jsonObject(with: data)) as? [String: Any],
              let type = command["type"] as? String else {
            emit(["type": "failed", "code": "WORKER_INVALID_JSON", "message": "invalid command"])
            return true
        }
        switch type {
        case "extract":
            let jobID = command["jobId"] as? String ?? ""
            let videoURL = command["videoUrl"] as? String ?? ""
            let duration = (command["durationSeconds"] as? NSNumber)?.doubleValue ?? .nan
            let request: ExtractRequest
            do {
                request = try validateExtractRequest(jobID: jobID, videoURL: videoURL, duration: duration)
            } catch {
                let code: String
                switch error {
                case ExtractValidationError.invalidJob: code = "OCR_INVALID_JOB"
                case ExtractValidationError.videoNotAllowed: code = "OCR_VIDEO_NOT_ALLOWED"
                default: code = "OCR_DURATION_NOT_SUPPORTED"
                }
                emit(["type": "failed", "jobId": jobID, "code": code, "message": code])
                emit(["type": "idle", "jobId": jobID])
                return true
            }
            stateLock.lock()
            if activeJob != nil {
                stateLock.unlock()
                emit(["type": "failed", "jobId": jobID, "code": "OCR_WORKER_BUSY", "message": "worker is busy"])
                emit(["type": "idle", "jobId": jobID])
                return true
            }
            activeJob = jobID
            cancelled = false
            stateLock.unlock()
            queue.async { self.process(request) }
        case "cancel":
            stateLock.lock()
            if activeJob == command["jobId"] as? String { cancelled = true }
            stateLock.unlock()
        case "shutdown":
            stateLock.lock()
            cancelled = true
            stateLock.unlock()
            return false
        default:
            emit(["type": "failed", "code": "WORKER_INVALID_COMMAND", "message": "unknown command"])
        }
        return true
    }

    func waitForIdle() { queue.sync {} }

    private func isCancelled() -> Bool {
        stateLock.lock()
        defer { stateLock.unlock() }
        return cancelled
    }

    private func process(_ request: ExtractRequest) {
        defer {
            stateLock.lock()
            activeJob = nil
            cancelled = false
            stateLock.unlock()
            emit(["type": "idle", "jobId": request.jobID])
        }
        let asset = AVURLAsset(url: request.videoURL)
        let generator = AVAssetImageGenerator(asset: asset)
        generator.appliesPreferredTrackTransform = true
        generator.requestedTimeToleranceBefore = .zero
        generator.requestedTimeToleranceAfter = .zero
        let total = max(1, Int(ceil(request.duration)))
        var body = ""
        var previous = ""
        var frames = 0
        var duplicates = 0
        do {
            for second in 0..<total {
                if isCancelled() { throw WorkerFailure(code: "OCR_CANCELLED", message: "task cancelled") }
                let image: CGImage
                do {
                    image = try generator.copyCGImage(at: CMTime(seconds: Double(second), preferredTimescale: 600), actualTime: nil)
                } catch {
                    emitProgress(request.jobID, second + 1, total)
                    continue
                }
                frames += 1
                let text = filterFrameLines(try recognize(image), seconds: second)
                if !text.isEmpty {
                    let merged = appendScrollText(body: body, previous: previous, current: text)
                    body = merged.text
                    previous = text
                    if merged.duplicate { duplicates += 1 }
                }
                emitProgress(request.jobID, second + 1, total)
            }
            if isCancelled() { throw WorkerFailure(code: "OCR_CANCELLED", message: "task cancelled") }
            body = cleanText(body)
            if frames == 0 { throw WorkerFailure(code: "OCR_VIDEO_READ_FAILED", message: "no video frames could be read") }
            if body.isEmpty { throw WorkerFailure(code: "OCR_NO_TEXT", message: "no text was recognized") }
            if body.utf8.count > 2 * 1024 * 1024 { throw WorkerFailure(code: "OCR_OUTPUT_TOO_LARGE", message: "OCR result exceeds 2 MiB") }
            let characters = body.filter { !$0.isWhitespace && !$0.isNewline }.count
            emit(["type": "complete", "jobId": request.jobID, "text": body, "characters": characters, "frames": frames, "duplicates": duplicates])
        } catch let failure as WorkerFailure {
            emit(["type": "failed", "jobId": request.jobID, "code": failure.code, "message": failure.message])
        } catch {
            emit(["type": "failed", "jobId": request.jobID, "code": "OCR_EXECUTION_FAILED", "message": "Vision OCR failed"])
        }
    }

    private func emitProgress(_ jobID: String, _ completed: Int, _ total: Int) {
        emit(["type": "progress", "jobId": jobID, "completed": completed, "total": total, "percent": completed * 100 / total])
    }

    private func recognize(_ image: CGImage) throws -> [FrameLine] {
        let request = VNRecognizeTextRequest()
        request.recognitionLevel = .accurate
        request.recognitionLanguages = ["zh-Hans", "en-US"]
        request.usesLanguageCorrection = true
        try VNImageRequestHandler(cgImage: image).perform([request])
        return (request.results ?? [])
            .sorted { $0.boundingBox.midY > $1.boundingBox.midY }
            .compactMap { observation -> FrameLine? in
                guard let text = observation.topCandidates(1).first?.string.trimmingCharacters(in: .whitespacesAndNewlines),
                      !text.isEmpty else { return nil }
                return FrameLine(text: text, y: observation.boundingBox.minY, height: observation.boundingBox.height)
            }
    }
}

private struct WorkerFailure: Error {
    let code: String
    let message: String
}

private let worker = ResidentWorker()
worker.emit(["type": "ready", "modelVersion": "macos-vision-v1"])
while let line = readLine() {
    if !worker.handle(line) { break }
}
worker.waitForIdle()
