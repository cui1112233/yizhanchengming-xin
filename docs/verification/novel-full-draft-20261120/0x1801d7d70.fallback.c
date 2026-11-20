/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x1801d7d70 */

longlong FUN_1801d7d70(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  undefined8 *puVar2;
  int *piVar3;
  longlong *plVar4;
  longlong *plVar5;
  longlong lVar6;
  code *pcVar7;
  code *pcVar8;
  undefined8 uVar9;

  plVar4 = DAT_1803bc8a0;
  plVar1 = (longlong *)*param_3;
  if (DAT_1803bc8a0 == (longlong *)0x0) {
LAB_1801d7dbc:
    DAT_1803bc8a0 = (longlong *)FUN_18036cd10(param_1,DAT_1803bc890,DAT_1803ca5e8,8);
  }
  else if ((1 < *DAT_1803bc8a0) || (DAT_1803bc8a0[2] != 0)) {
    if ((DAT_1803bc8a0 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803bc8a0 && (*DAT_1803bc8a0 = *DAT_1803bc8a0 + -1, *plVar4 == 0)))) {
      (**(code **)(plVar4[1] + 0x30))(plVar4);
    }
    goto LAB_1801d7dbc;
  }
  plVar4 = DAT_1803bc8a0;
  FUN_180001d00(param_1,DAT_1803bc8a0);
  uVar9 = DAT_1803bc9e0;
  piVar3 = DAT_1803bc970;
  plVar5 = (longlong *)FUN_180372160(param_1,0xc);
  if (*piVar3 + 1 != 0) {
    *piVar3 = *piVar3 + 1;
  }
  puVar2 = (undefined8 *)plVar5[3];
  *puVar2 = piVar3;
  lVar6 = FUN_18037f0c0(puVar2,plVar1,DAT_1803bc978);
  if (lVar6 == 0) {
    uVar9 = 0x106;
LAB_1801d7eee:
    lVar6 = *(longlong *)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
    if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
      (**(code **)(plVar5[1] + 0x30))(plVar5);
    }
  }
  else {
    *(longlong *)(plVar5[3] + 8) = lVar6;
    piVar3 = DAT_1803bc980;
    if (*DAT_1803bc980 + 1 != 0) {
      *DAT_1803bc980 = *DAT_1803bc980 + 1;
    }
    *(int **)(plVar5[3] + 0x10) = piVar3;
    piVar3 = DAT_1803bcb80;
    if (*DAT_1803bcb80 + 1 != 0) {
      *DAT_1803bcb80 = *DAT_1803bcb80 + 1;
    }
    *(int **)(plVar5[3] + 0x18) = piVar3;
    lVar6 = FUN_18037f0c0(piVar3,plVar1,DAT_1803bcaf0);
    if (lVar6 == 0) {
      uVar9 = 0x109;
      goto LAB_1801d7eee;
    }
    *(longlong *)(plVar5[3] + 0x20) = lVar6;
    piVar3 = DAT_1803bc980;
    if (*DAT_1803bc980 + 1 != 0) {
      *DAT_1803bc980 = *DAT_1803bc980 + 1;
    }
    *(int **)(plVar5[3] + 0x28) = piVar3;
    piVar3 = DAT_1803bcb88;
    if (*DAT_1803bcb88 + 1 != 0) {
      *DAT_1803bcb88 = *DAT_1803bcb88 + 1;
    }
    *(int **)(plVar5[3] + 0x30) = piVar3;
    lVar6 = FUN_18037f0c0(piVar3,plVar1,DAT_1803bcb90);
    if (lVar6 == 0) {
      uVar9 = 0x10c;
      goto LAB_1801d7eee;
    }
    *(longlong *)(plVar5[3] + 0x38) = lVar6;
    piVar3 = DAT_1803bc980;
    if (*DAT_1803bc980 + 1 != 0) {
      *DAT_1803bc980 = *DAT_1803bc980 + 1;
    }
    *(int **)(plVar5[3] + 0x40) = piVar3;
    piVar3 = DAT_1803bcb98;
    if (*DAT_1803bcb98 + 1 != 0) {
      *DAT_1803bcb98 = *DAT_1803bcb98 + 1;
    }
    *(int **)(plVar5[3] + 0x48) = piVar3;
    piVar3 = DAT_1803bcba0;
    if (*DAT_1803bcba0 + 1 != 0) {
      *DAT_1803bcba0 = *DAT_1803bcba0 + 1;
    }
    *(int **)(plVar5[3] + 0x50) = piVar3;
    piVar3 = DAT_1803bcba8;
    if (*DAT_1803bcba8 + 1 != 0) {
      *DAT_1803bcba8 = *DAT_1803bcba8 + 1;
    }
    *(int **)(plVar5[3] + 0x58) = piVar3;
    lVar6 = PyUnicode_Join(uVar9,plVar5);
    if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
      (**(code **)(plVar5[1] + 0x30))(plVar5);
    }
    if (lVar6 != 0) {
      FUN_180001d80(param_1);
      if ((int)*plVar1 < 0) {
        return lVar6;
      }
      *plVar1 = *plVar1 + -1;
      if (*plVar1 != 0) {
        return lVar6;
      }
      (**(code **)(plVar1[1] + 0x30))(plVar1);
      return lVar6;
    }
    lVar6 = *(longlong *)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
    uVar9 = 0x104;
  }
  pcVar8 = (code *)0x0;
  piVar3 = *(int **)(lVar6 + 0x28);
  if (piVar3 == (int *)0x0) {
    pcVar7 = (code *)FUN_18038e2a0(plVar4,uVar9);
    plVar5 = *(longlong **)(lVar6 + 0x28);
    if (pcVar7 != _Py_NoneStruct_exref) {
LAB_1801d802c:
      pcVar8 = pcVar7;
      pcVar7 = pcVar8;
      if ((pcVar8 != (code *)0x0) && (*(int *)pcVar8 + 1 != 0)) {
        *(int *)pcVar8 = *(int *)pcVar8 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar3 + 6) == plVar4) goto LAB_1801d8075;
    pcVar7 = (code *)FUN_18038e2a0(plVar4,uVar9);
    *(int **)(pcVar7 + 0x10) = piVar3;
    if (*piVar3 + 1 != 0) {
      *piVar3 = *piVar3 + 1;
    }
    plVar5 = *(longlong **)(lVar6 + 0x28);
    if (pcVar7 != _Py_NoneStruct_exref) goto LAB_1801d802c;
  }
  *(code **)(lVar6 + 0x28) = pcVar8;
  if (((plVar5 != (longlong *)0x0) && (-1 < (int)*plVar5)) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)
     ) {
    (**(code **)(plVar5[1] + 0x30))();
  }
  if (((pcVar7 != (code *)0x0) && (-1 < *(int *)pcVar7)) &&
     (*(longlong *)pcVar7 = *(longlong *)pcVar7 + -1, *(longlong *)pcVar7 == 0)) {
    (**(code **)(*(longlong *)(pcVar7 + 8) + 0x30))(pcVar7);
  }
LAB_1801d8075:
  FUN_18036d0a0(plVar4,&DAT_180394940,plVar1);
  plVar5 = DAT_1803bc8a0;
  if (plVar4 == DAT_1803bc8a0) {
    if ((-1 < (int)*DAT_1803bc8a0) && (*DAT_1803bc8a0 = *DAT_1803bc8a0 + -1, *plVar5 == 0)) {
      (**(code **)(plVar5[1] + 0x30))();
    }
    DAT_1803bc8a0 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))(plVar1);
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong *)(param_1 + 0x60) = lVar6;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return 0;
}
