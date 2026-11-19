/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x1803839a0 */

undefined8 FUN_1803839a0(longlong param_1,longlong param_2)

{
  int *piVar1;
  ulonglong uVar2;
  longlong lVar3;
  
  if (param_1 == param_2) {
    return 1;
  }
  uVar2 = *(ulonglong *)(param_1 + 0x10) >> 3;
  if ((1 - (ulonglong)((uint)*(ulonglong *)(param_1 + 0x10) & 3)) * uVar2 -
      (1 - (ulonglong)((uint)*(ulonglong *)(param_2 + 0x10) & 3)) *
      (*(ulonglong *)(param_2 + 0x10) >> 3) != 0) {
    return 0;
  }
  lVar3 = uVar2 - 1;
  if (-1 < lVar3) {
    piVar1 = (int *)(param_1 + 0x18 + lVar3 * 4);
    do {
      if (*piVar1 != *(int *)((longlong)piVar1 + (param_2 - param_1))) {
        return 0;
      }
      piVar1 = piVar1 + -1;
      lVar3 = lVar3 + -1;
    } while (-1 < lVar3);
  }
  return 1;
}
