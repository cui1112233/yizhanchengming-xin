Novel document parser notices

The application links github.com/klippa-app/go-pdfium v1.21.4 (MIT), including
its default embedded PDFium WASM, and github.com/tetratelabs/wazero v1.12.0
(Apache-2.0). The WASM is unchanged. There is no external PDF executable.

Wrapper/runtime licenses were copied from the exact Go module versions.
PDFium and third-party notices in pdfium-dependencies were obtained from the
upstream chromium/8111 pdfium-wasm.tgz distribution:
https://github.com/bblanchon/pdfium-binaries/releases/download/chromium/8111/pdfium-wasm.tgz
The wrapper README names PDFium 8111; its NATIVE_VS_WEBASSEMBLY.md documents
the bundled WASM patch provenance (jerbob92/pdfium-binaries wasm-standalone).
These notices cover upstream components, without asserting the two binaries
are identical. Additional PDFium main LICENSE is retained for completeness.

Source and corresponding notices:
https://github.com/klippa-app/go-pdfium/tree/v1.21.4
https://github.com/jerbob92/pdfium-binaries/tree/wasm-standalone
https://pdfium.googlesource.com/pdfium/
https://github.com/tetratelabs/wazero/tree/v1.12.0

The licenses directory is embedded in the server and available through
GET /api/novel-creation/documents/licenses (authenticated).
