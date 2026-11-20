/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x1800d6450 */

/* WARNING: Function: __security_check_cookie replaced with injection: security_check_cookie */

longlong FUN_1800d6450(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  longlong *plVar2;
  longlong *plVar3;
  longlong *plVar4;
  int iVar5;
  longlong *plVar6;
  longlong *plVar7;
  undefined8 uVar8;
  longlong lVar9;
  code *pcVar10;
  code *pcVar11;
  undefined1 auStack_98 [32];
  longlong *local_78;
  int *local_68;
  longlong *local_60;
  undefined8 local_58;
  longlong *local_50;
  ulonglong local_48;

  plVar4 = DAT_1803b2a90;
  local_48 = DAT_1803a6040 ^ (ulonglong)auStack_98;
  plVar1 = (longlong *)*param_3;
  plVar2 = (longlong *)param_3[1];
  plVar3 = (longlong *)param_3[2];
  if (DAT_1803b2a90 == (longlong *)0x0) {
LAB_1800d64b0:
    DAT_1803b2a90 = (longlong *)FUN_18036cd10(param_1,DAT_1803b2908,DAT_1803ca7a0,0x18);
  }
  else if ((1 < *DAT_1803b2a90) || (DAT_1803b2a90[2] != 0)) {
    if ((DAT_1803b2a90 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b2a90 && (*DAT_1803b2a90 = *DAT_1803b2a90 + -1, *plVar4 == 0)))) {
      (**(code **)(plVar4[1] + 0x30))(plVar4);
    }
    goto LAB_1800d64b0;
  }
  plVar4 = DAT_1803b2a90;
  FUN_180001d00(param_1,DAT_1803b2a90);
  plVar6 = (longlong *)FUN_18037f0c0();
  if (plVar6 == (longlong *)0x0) {
LAB_1800d6601:
    lVar9 = *(longlong *)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
  }
  else {
    plVar7 = (longlong *)FUN_18037f0c0();
    if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
      (**(code **)(plVar6[1] + 0x30))(plVar6);
    }
    local_68 = (int *)DAT_1803b2c08;
    if (plVar7 == (longlong *)0x0) goto LAB_1800d6601;
    *(undefined4 *)(plVar4 + 5) = 0x1ea;
    plVar6 = (longlong *)FUN_18037e6c0(param_1,plVar1,DAT_1803b31d0,plVar2);
    if (plVar6 != (longlong *)0x0) {
      uVar8 = _PyDict_NewPresized(1);
      PyDict_SetItem(uVar8,local_68,plVar6);
      if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
        (**(code **)(plVar6[1] + 0x30))(plVar6);
      }
      iVar5 = (int)*plVar3 + 1;
      if (iVar5 != 0) {
        *(int *)plVar3 = iVar5;
      }
      local_60 = plVar7;
      local_58 = uVar8;
      local_50 = plVar3;
      lVar9 = FUN_180097770(param_1,&local_60);
      if (lVar9 != 0) {
        FUN_180001d80(param_1);
        if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
          (**(code **)(plVar1[1] + 0x30))(plVar1);
        }
        if ((-1 < (int)*plVar2) && (*plVar2 = *plVar2 + -1, *plVar2 == 0)) {
          (**(code **)(plVar2[1] + 0x30))(plVar2);
        }
        if ((int)*plVar3 < 0) {
          return lVar9;
        }
        *plVar3 = *plVar3 + -1;
        if (*plVar3 != 0) {
          return lVar9;
        }
        (**(code **)(plVar3[1] + 0x30))(plVar3);
        return lVar9;
      }
      goto LAB_1800d6601;
    }
    lVar9 = *(longlong *)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
    if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
      (**(code **)(plVar7[1] + 0x30))(plVar7);
    }
  }
  pcVar11 = (code *)0x0;
  local_68 = *(int **)(lVar9 + 0x28);
  if (local_68 == (int *)0x0) {
    pcVar10 = (code *)FUN_18038e2a0(plVar4,0x1ea);
    plVar6 = *(longlong **)(lVar9 + 0x28);
    if (((pcVar10 != _Py_NoneStruct_exref) && (pcVar11 = pcVar10, pcVar10 != (code *)0x0)) &&
       (*(int *)pcVar10 + 1 != 0)) {
      *(int *)pcVar10 = *(int *)pcVar10 + 1;
    }
  }
  else {
    if (*(longlong **)(local_68 + 6) == plVar4) goto LAB_1800d66d9;
    pcVar10 = (code *)FUN_18038e2a0(plVar4,0x1ea);
    *(int **)(pcVar10 + 0x10) = local_68;
    if (*local_68 + 1 != 0) {
      *local_68 = *local_68 + 1;
    }
    plVar6 = *(longlong **)(lVar9 + 0x28);
    if (((pcVar10 != _Py_NoneStruct_exref) && (pcVar11 = pcVar10, pcVar10 != (code *)0x0)) &&
       (*(int *)pcVar10 + 1 != 0)) {
      *(int *)pcVar10 = *(int *)pcVar10 + 1;
    }
  }
  *(code **)(lVar9 + 0x28) = pcVar11;
  if (((plVar6 != (longlong *)0x0) && (-1 < (int)*plVar6)) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)
     ) {
    (**(code **)(plVar6[1] + 0x30))();
  }
  if (((pcVar10 != (code *)0x0) && (-1 < *(int *)pcVar10)) &&
     (*(longlong *)pcVar10 = *(longlong *)pcVar10 + -1, *(longlong *)pcVar10 == 0)) {
    (**(code **)(*(longlong *)(pcVar10 + 8) + 0x30))(pcVar10);
  }
LAB_1800d66d9:
  local_78 = plVar3;
  FUN_18036d0a0(plVar4,&DAT_1803948e8,plVar1,plVar2);
  plVar6 = DAT_1803b2a90;
  if (plVar4 == DAT_1803b2a90) {
    if ((-1 < (int)*DAT_1803b2a90) && (*DAT_1803b2a90 = *DAT_1803b2a90 + -1, *plVar6 == 0)) {
      (**(code **)(plVar6[1] + 0x30))();
    }
    DAT_1803b2a90 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))(plVar1);
  }
  if ((-1 < (int)*plVar2) && (*plVar2 = *plVar2 + -1, *plVar2 == 0)) {
    (**(code **)(plVar2[1] + 0x30))(plVar2);
  }
  if ((-1 < (int)*plVar3) && (*plVar3 = *plVar3 + -1, *plVar3 == 0)) {
    (**(code **)(plVar3[1] + 0x30))(plVar3);
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong *)(param_1 + 0x60) = lVar9;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return 0;
}
