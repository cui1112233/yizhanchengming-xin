/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x1802c6ac0 */

/* WARNING: Function: __security_check_cookie replaced with injection: security_check_cookie */

longlong FUN_1802c6ac0(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  int *piVar2;
  longlong *plVar3;
  undefined8 uVar4;
  longlong lVar5;
  longlong *plVar6;
  longlong *plVar7;
  code *pcVar8;
  code *pcVar9;
  undefined1 auStack_78 [40];
  longlong local_50;
  undefined8 local_48;
  undefined8 local_40;
  longlong *local_38;
  ulonglong local_30;
  
  plVar3 = DAT_1803c4b98;
  local_30 = DAT_1803a6040 ^ (ulonglong)auStack_78;
  plVar1 = (longlong *)*param_3;
  local_50 = DAT_180391e98;
  if (DAT_1803c4b98 == (longlong *)0x0) {
LAB_1802c6b21:
    DAT_1803c4b98 = (longlong *)FUN_18036cd10(param_1,DAT_1803c4bd8,DAT_1803ca468,8);
  }
  else if ((1 < *DAT_1803c4b98) || (DAT_1803c4b98[2] != 0)) {
    if ((DAT_1803c4b98 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803c4b98 && (*DAT_1803c4b98 = *DAT_1803c4b98 + -1, *plVar3 == 0)))) {
      (**(code **)(plVar3[1] + 0x30))(plVar3);
    }
    goto LAB_1802c6b21;
  }
  plVar3 = DAT_1803c4b98;
  FUN_180001d00(param_1,DAT_1803c4b98);
  lVar5 = FUN_1802c23b0(param_1);
  uVar4 = DAT_1803c4d48;
  pcVar9 = (code *)0x0;
  if (lVar5 == 0) {
    FUN_180373340(param_1,&local_50,DAT_1803c4c50);
    lVar5 = local_50;
  }
  else {
    local_40 = DAT_1803c4d50;
    *(undefined4 *)(plVar3 + 5) = 0x6a;
    local_48 = uVar4;
    local_38 = plVar1;
    plVar6 = (longlong *)FUN_18037e980(param_1,lVar5,DAT_1803c4d40,&local_48);
    if (plVar6 != (longlong *)0x0) {
      plVar7 = (longlong *)FUN_18037f0c0();
      if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
        (**(code **)(plVar6[1] + 0x30))(plVar6);
      }
      if (plVar7 != (longlong *)0x0) {
        *(undefined4 *)(plVar3 + 5) = 0x6a;
        lVar5 = FUN_180376210(param_1,plVar7);
        if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
          (**(code **)(plVar7[1] + 0x30))(plVar7);
        }
        if (lVar5 != 0) {
          FUN_180001d80(param_1);
          if ((int)*plVar1 < 0) {
            return lVar5;
          }
          *plVar1 = *plVar1 + -1;
          if (*plVar1 != 0) {
            return lVar5;
          }
          (**(code **)(plVar1[1] + 0x30))(plVar1);
          return lVar5;
        }
      }
    }
    lVar5 = *(longlong *)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
  }
  piVar2 = *(int **)(lVar5 + 0x28);
  if (piVar2 == (int *)0x0) {
    pcVar8 = (code *)FUN_18038e2a0(plVar3,0x6a);
    plVar6 = *(longlong **)(lVar5 + 0x28);
    if (pcVar8 != _Py_NoneStruct_exref) {
LAB_1802c6ca7:
      pcVar9 = pcVar8;
      pcVar8 = pcVar9;
      if ((pcVar9 != (code *)0x0) && (*(int *)pcVar9 + 1 != 0)) {
        *(int *)pcVar9 = *(int *)pcVar9 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar2 + 6) == plVar3) goto LAB_1802c6cf0;
    pcVar8 = (code *)FUN_18038e2a0(plVar3,0x6a);
    *(int **)(pcVar8 + 0x10) = piVar2;
    if (*piVar2 + 1 != 0) {
      *piVar2 = *piVar2 + 1;
    }
    plVar6 = *(longlong **)(lVar5 + 0x28);
    if (pcVar8 != _Py_NoneStruct_exref) goto LAB_1802c6ca7;
  }
  *(code **)(lVar5 + 0x28) = pcVar9;
  if (((plVar6 != (longlong *)0x0) && (-1 < (int)*plVar6)) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)
     ) {
    (**(code **)(plVar6[1] + 0x30))();
  }
  if (((pcVar8 != (code *)0x0) && (-1 < *(int *)pcVar8)) &&
     (*(longlong *)pcVar8 = *(longlong *)pcVar8 + -1, *(longlong *)pcVar8 == 0)) {
    (**(code **)(*(longlong *)(pcVar8 + 8) + 0x30))(pcVar8);
  }
LAB_1802c6cf0:
  FUN_18036d0a0(plVar3,&DAT_180394940,plVar1);
  plVar6 = DAT_1803c4b98;
  if (plVar3 == DAT_1803c4b98) {
    if ((-1 < (int)*DAT_1803c4b98) && (*DAT_1803c4b98 = *DAT_1803c4b98 + -1, *plVar6 == 0)) {
      (**(code **)(plVar6[1] + 0x30))();
    }
    DAT_1803c4b98 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))(plVar1);
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong *)(param_1 + 0x60) = lVar5;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return 0;
}

