/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x180375370 */

undefined8 FUN_180375370(longlong param_1,undefined8 param_2)

{
  longlong lVar1;
  undefined8 uVar2;

  lVar1 = FUN_180373860(param_2);
  if ((lVar1 < 0) && (*(longlong *)(param_1 + 0x60) != 0)) {
    return 0;
  }
                    /* WARNING: Could not recover jumptable at 0x00018037539d. Too many branches */
                    /* WARNING: Treating indirect jump as call */
  uVar2 = PyLong_FromSsize_t(lVar1);
  return uVar2;
}
