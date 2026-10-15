// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "GiantMaterialOCRWorker",
    platforms: [.macOS(.v13)],
    products: [.executable(name: "GiantMaterialOCRWorker", targets: ["GiantMaterialOCRWorker"])],
    targets: [
        .target(name: "GiantOCRCore", path: "Sources/GiantOCRCore"),
        .executableTarget(name: "GiantMaterialOCRWorker", dependencies: ["GiantOCRCore"], path: "Sources/GiantMaterialOCRWorker")
    ]
)
