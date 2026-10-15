!ifndef PRODUCT_VERSION
  !define PRODUCT_VERSION "0.1.0"
!endif
!ifndef INPUT_STAGE
  !define INPUT_STAGE "stage"
!endif
!ifndef OUTPUT_FILE
  !define OUTPUT_FILE "GiantMaterialExecutor-${PRODUCT_VERSION}.exe"
!endif

Name "Giant Material Executor ${PRODUCT_VERSION}"
OutFile "${OUTPUT_FILE}"
InstallDir "$LOCALAPPDATA\YizhanShengming\GiantMaterialExecutor"
RequestExecutionLevel user
Unicode True

VIProductVersion "${PRODUCT_VERSION}.0"
VIAddVersionKey "ProductName" "Giant Material Executor"
VIAddVersionKey "FileDescription" "Resident Windows OCR executor"
VIAddVersionKey "ProductVersion" "${PRODUCT_VERSION}"

Section "Install"
  SetOutPath "$INSTDIR"
  File "${INPUT_STAGE}\GiantMaterialExecutor.exe"
  SetOutPath "$INSTDIR\worker"
  File "${INPUT_STAGE}\worker\ocr_worker.py"
  File "${INPUT_STAGE}\worker\requirements-lock.txt"
  CreateDirectory "$SMPROGRAMS\Yizhan Shengming"
  CreateShortCut "$SMPROGRAMS\Yizhan Shengming\Giant Material Executor.lnk" "$INSTDIR\GiantMaterialExecutor.exe"
  WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Run" "GiantMaterialExecutor" '"$INSTDIR\GiantMaterialExecutor.exe"'
SectionEnd

Section "Uninstall"
  Delete "$SMPROGRAMS\Yizhan Shengming\Giant Material Executor.lnk"
  RMDir "$SMPROGRAMS\Yizhan Shengming"
  DeleteRegValue HKCU "Software\Microsoft\Windows\CurrentVersion\Run" "GiantMaterialExecutor"
  Delete "$INSTDIR\GiantMaterialExecutor.exe"
  Delete "$INSTDIR\worker\ocr_worker.py"
  Delete "$INSTDIR\worker\requirements-lock.txt"
  RMDir "$INSTDIR\worker"
  RMDir "$INSTDIR"
SectionEnd
