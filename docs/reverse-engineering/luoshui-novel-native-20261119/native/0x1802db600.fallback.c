/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x1802db600 */

longlong * FUN_1802db600(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  longlong *plVar2;
  longlong *plVar3;
  longlong lVar4;
  int *piVar5;
  longlong lVar6;
  longlong *plVar7;
  int iVar8;
  uint uVar9;
  code *pcVar10;
  code *pcVar11;
  longlong *plVar12;
  undefined8 uVar13;
  
  plVar7 = DAT_1803c5370;
  plVar1 = (longlong *)*param_3;
  plVar2 = (longlong *)param_3[1];
  plVar3 = (longlong *)param_3[2];
  if (DAT_1803c5370 == (longlong *)0x0) {
LAB_1802db650:
    DAT_1803c5370 = (longlong *)FUN_18036cd10(param_1,DAT_1803c5460,DAT_1803ca450,0x18);
  }
  else if ((1 < *DAT_1803c5370) || (DAT_1803c5370[2] != 0)) {
    if ((DAT_1803c5370 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803c5370 && (*DAT_1803c5370 = *DAT_1803c5370 + -1, *plVar7 == 0)))) {
      (**(code **)(plVar7[1] + 0x30))(plVar7);
    }
    goto LAB_1802db650;
  }
  plVar7 = DAT_1803c5370;
  FUN_180001d00(param_1,DAT_1803c5370);
  iVar8 = FUN_1803836d0(plVar1,DAT_1803c5648);
  if (iVar8 != -1) {
    plVar12 = DAT_1803c5648;
    if (iVar8 != 1) {
      iVar8 = FUN_1803836d0(plVar2,DAT_1803c5648);
      if (iVar8 == -1) {
        uVar13 = 0xa4;
        goto LAB_1802db741;
      }
      plVar12 = DAT_1803c5648;
      if (iVar8 != 1) {
        iVar8 = FUN_180001510(plVar3);
        if ((iVar8 == -1) ||
           ((uVar9 = (uint)(iVar8 != 0), iVar8 == 0 &&
            ((iVar8 = FUN_1803836d0(plVar1,DAT_1803c5718), iVar8 == -1 ||
             ((uVar9 = 1, iVar8 != 1 &&
              (uVar9 = FUN_1803836d0(plVar2,DAT_1803c5718), uVar9 == 0xffffffff)))))))) {
          uVar13 = 0xa6;
          goto LAB_1802db741;
        }
        plVar12 = DAT_1803c5718;
        if (uVar9 != 1) {
          FUN_180001d80(param_1);
          iVar8 = (int)*plVar1 + 1;
          plVar12 = plVar1;
          if (iVar8 != 0) {
            *(int *)plVar1 = iVar8;
          }
          goto LAB_1802db94a;
        }
      }
    }
    iVar8 = (int)*plVar12 + 1;
    if (iVar8 != 0) {
      *(int *)plVar12 = iVar8;
    }
    FUN_180001d80(param_1);
LAB_1802db94a:
    if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
      (**(code **)(plVar1[1] + 0x30))(plVar1);
    }
    if ((-1 < (int)*plVar2) && (*plVar2 = *plVar2 + -1, *plVar2 == 0)) {
      (**(code **)(plVar2[1] + 0x30))(plVar2);
    }
    if ((int)*plVar3 < 0) {
      return plVar12;
    }
    *plVar3 = *plVar3 + -1;
    if (*plVar3 != 0) {
      return plVar12;
    }
    (**(code **)(plVar3[1] + 0x30))(plVar3);
    return plVar12;
  }
  uVar13 = 0xa4;
LAB_1802db741:
  pcVar11 = (code *)0x0;
  lVar4 = *(longlong *)(param_1 + 0x60);
  *(undefined8 *)(param_1 + 0x60) = 0;
  piVar5 = *(int **)(lVar4 + 0x28);
  if (piVar5 == (int *)0x0) {
    pcVar10 = (code *)FUN_18038e2a0(plVar7,uVar13);
    plVar12 = *(longlong **)(lVar4 + 0x28);
    if (((pcVar10 != _Py_NoneStruct_exref) && (pcVar11 = pcVar10, pcVar10 != (code *)0x0)) &&
       (*(int *)pcVar10 + 1 != 0)) {
      *(int *)pcVar10 = *(int *)pcVar10 + 1;
    }
  }
  else {
    if (*(longlong **)(piVar5 + 6) == plVar7) goto LAB_1802db80d;
    pcVar10 = (code *)FUN_18038e2a0(plVar7,uVar13);
    *(int **)(pcVar10 + 0x10) = piVar5;
    if (*piVar5 + 1 != 0) {
      *piVar5 = *piVar5 + 1;
    }
    plVar12 = *(longlong **)(lVar4 + 0x28);
    if (((pcVar10 != _Py_NoneStruct_exref) && (pcVar11 = pcVar10, pcVar10 != (code *)0x0)) &&
       (*(int *)pcVar10 + 1 != 0)) {
      *(int *)pcVar10 = *(int *)pcVar10 + 1;
    }
  }
  *(code **)(lVar4 + 0x28) = pcVar11;
  if (((plVar12 != (longlong *)0x0) && (-1 < (int)*plVar12)) &&
     (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
    (**(code **)(plVar12[1] + 0x30))();
  }
  if (((pcVar10 != (code *)0x0) && (-1 < *(int *)pcVar10)) &&
     (*(longlong *)pcVar10 = *(longlong *)pcVar10 + -1, *(longlong *)pcVar10 == 0)) {
    (**(code **)(*(longlong *)(pcVar10 + 8) + 0x30))(pcVar10);
  }
LAB_1802db80d:
  FUN_18036d0a0(plVar7,&DAT_1803948e8,plVar1,plVar2,plVar3);
  plVar12 = DAT_1803c5370;
  if (plVar7 == DAT_1803c5370) {
    if ((-1 < (int)*DAT_1803c5370) && (*DAT_1803c5370 = *DAT_1803c5370 + -1, *plVar12 == 0)) {
      (**(code **)(plVar12[1] + 0x30))();
    }
    DAT_1803c5370 = (longlong *)0x0;
  }
  lVar6 = **(longlong **)(param_1 + 0x38);
  plVar7 = *(longlong **)(lVar6 + 0x30);
  **(longlong **)(param_1 + 0x38) = *(longlong *)(lVar6 + 8);
  *(undefined4 *)(plVar7 + 8) = 0xffffffff;
  plVar12 = (longlong *)plVar7[2];
  if (((plVar12 != (longlong *)0x0) && (plVar7[2] = 0, -1 < (int)*plVar12)) &&
     (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
    (**(code **)(plVar12[1] + 0x30))();
  }
  if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
    (**(code **)(plVar7[1] + 0x30))(plVar7);
  }
  plVar7[10] = 0;
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
  *(longlong *)(param_1 + 0x60) = lVar4;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return (longlong *)0x0;
}

