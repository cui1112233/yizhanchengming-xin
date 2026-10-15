import Foundation

public enum ExtractValidationError: Error {
    case invalidJob
    case videoNotAllowed
    case durationNotSupported
}

public struct ExtractRequest {
    public let jobID: String
    public let videoURL: URL
    public let duration: Double
}

public func validateExtractRequest(jobID: String, videoURL: String, duration: Double) throws -> ExtractRequest {
    let jobID = jobID.trimmingCharacters(in: .whitespacesAndNewlines)
    guard !jobID.isEmpty else { throw ExtractValidationError.invalidJob }
    guard let url = URL(string: videoURL),
          url.scheme == "https",
          ["material.hnqingyuwen.top", "mlzr-material.hnqingyuwen.top", "ml-material.hnqingyuwen.top"].contains(url.host),
          url.port == nil,
          url.user == nil,
          url.password == nil,
          !url.path.isEmpty else { throw ExtractValidationError.videoNotAllowed }
    guard duration.isFinite, duration > 0, duration <= 1800 else { throw ExtractValidationError.durationNotSupported }
    return ExtractRequest(jobID: jobID, videoURL: url, duration: duration)
}
