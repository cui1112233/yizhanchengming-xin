Synthetic test fixtures created locally for this change. No user documents.

unicode.pdf: one page containing two Chinese chapter headings and sample text,
generated with installed PyMuPDF using the built-in china-s font.
encrypted.pdf: the same sample encrypted with AES-256, owner/user passwords.
encrypted-empty-user.pdf: AES-256 with an empty user password and owner password.
scanned.pdf: raster-only rendering of the sample, with no text objects.

Go tests read these fixtures directly; Python/PyMuPDF is not a runtime or test
dependency. ASCII, large-page-count and oversized-text PDFs are built in Go.
