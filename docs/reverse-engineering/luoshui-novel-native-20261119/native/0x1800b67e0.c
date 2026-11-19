
code * FUN_1800b67e0(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  longlong *plVar2;
  code *pcVar3;
  longlong *plVar4;
  int *piVar5;
  longlong lVar6;
  longlong *plVar7;
  int iVar8;
  longlong *plVar9;
  longlong *plVar10;
  code *pcVar11;
  code *pcVar12;
  longlong lVar13;
  
  plVar7 = DAT_1803b1ea8;
  plVar1 = (longlong *)*param_3;
  plVar2 = (longlong *)param_3[1];
  pcVar3 = (code *)param_3[2];
  plVar4 = (longlong *)param_3[3];
  if (DAT_1803b1ea8 == (longlong *)0x0) {
LAB_1800b6840:
    DAT_1803b1ea8 = (longlong *)FUN_18036cd10(param_1,DAT_1803b2578,DAT_1803ca7c0,0x20);
  }
  else if ((1 < *DAT_1803b1ea8) || (DAT_1803b1ea8[2] != 0)) {
    if ((DAT_1803b1ea8 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b1ea8 && (*DAT_1803b1ea8 = *DAT_1803b1ea8 + -1, *plVar7 == 0)))) {
      (**(code **)(plVar7[1] + 0x30))(plVar7);
    }
    goto LAB_1800b6840;
  }
  plVar7 = DAT_1803b1ea8;
  FUN_180001d00(param_1,DAT_1803b1ea8);
  plVar9 = (longlong *)FUN_18037f0c0();
  if (plVar9 == (longlong *)0x0) {
    lVar13 = *(longlong *)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
  }
  else {
    *(undefined4 *)(plVar7 + 5) = 0x3b;
    plVar10 = (longlong *)FUN_1803770d0(param_1,plVar9,DAT_1803b1fa8);
    if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
      (**(code **)(plVar9[1] + 0x30))(plVar9);
    }
    if (plVar10 == (longlong *)0x0) {
      lVar13 = *(longlong *)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
    }
    else {
      iVar8 = FUN_180001510(plVar10);
      if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
        (**(code **)(plVar10[1] + 0x30))(plVar10);
      }
      if (iVar8 == -1) {
        lVar13 = *(longlong *)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
      }
      else {
        pcVar11 = _Py_FalseStruct_exref;
        if (iVar8 != 0) {
          pcVar11 = _Py_TrueStruct_exref;
        }
        iVar8 = FUN_180001510(pcVar11);
        if (iVar8 != 1) {
          if (*(int *)pcVar11 + 1 != 0) {
            *(int *)pcVar11 = *(int *)pcVar11 + 1;
          }
LAB_1800b6bd7:
          FUN_180001d80(param_1);
          if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
            (**(code **)(plVar1[1] + 0x30))(plVar1);
          }
          if ((-1 < (int)*plVar2) && (*plVar2 = *plVar2 + -1, *plVar2 == 0)) {
            (**(code **)(plVar2[1] + 0x30))(plVar2);
          }
          if ((-1 < *(int *)pcVar3) &&
             (*(longlong *)pcVar3 = *(longlong *)pcVar3 + -1, *(longlong *)pcVar3 == 0)) {
            (**(code **)(*(longlong *)(pcVar3 + 8) + 0x30))(pcVar3);
          }
          if ((int)*plVar4 < 0) {
            return pcVar11;
          }
          *plVar4 = *plVar4 + -1;
          if (*plVar4 != 0) {
            return pcVar11;
          }
          (**(code **)(plVar4[1] + 0x30))(plVar4);
          return pcVar11;
        }
        pcVar11 = (code *)FUN_180384d50(plVar2,plVar4);
        if (pcVar11 == (code *)0x0) {
          lVar13 = *(longlong *)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
        }
        else {
          iVar8 = FUN_180001510(pcVar11);
          if (iVar8 != -1) {
            if (iVar8 == 1) {
              if ((-1 < *(int *)pcVar11) &&
                 (*(longlong *)pcVar11 = *(longlong *)pcVar11 + -1, *(longlong *)pcVar11 == 0)) {
                (**(code **)(*(longlong *)(pcVar11 + 8) + 0x30))(pcVar11);
              }
              pcVar11 = pcVar3;
              if (*(int *)pcVar3 + 1 != 0) {
                *(int *)pcVar3 = *(int *)pcVar3 + 1;
              }
            }
            goto LAB_1800b6bd7;
          }
          lVar13 = *(longlong *)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          if ((-1 < *(int *)pcVar11) &&
             (*(longlong *)pcVar11 = *(longlong *)pcVar11 + -1, *(longlong *)pcVar11 == 0)) {
            (**(code **)(*(longlong *)(pcVar11 + 8) + 0x30))(pcVar11);
          }
        }
      }
    }
  }
  pcVar11 = (code *)0x0;
  piVar5 = *(int **)(lVar13 + 0x28);
  if (piVar5 == (int *)0x0) {
    pcVar12 = (code *)FUN_18038e2a0(plVar7,0x3b);
    plVar9 = *(longlong **)(lVar13 + 0x28);
    if (((pcVar12 != _Py_NoneStruct_exref) && (pcVar11 = pcVar12, pcVar12 != (code *)0x0)) &&
       (*(int *)pcVar12 + 1 != 0)) {
      *(int *)pcVar12 = *(int *)pcVar12 + 1;
    }
  }
  else {
    if (*(longlong **)(piVar5 + 6) == plVar7) goto LAB_1800b6a6a;
    pcVar12 = (code *)FUN_18038e2a0(plVar7,0x3b);
    *(int **)(pcVar12 + 0x10) = piVar5;
    if (*piVar5 + 1 != 0) {
      *piVar5 = *piVar5 + 1;
    }
    plVar9 = *(longlong **)(lVar13 + 0x28);
    if (((pcVar12 != _Py_NoneStruct_exref) && (pcVar11 = pcVar12, pcVar12 != (code *)0x0)) &&
       (*(int *)pcVar12 + 1 != 0)) {
      *(int *)pcVar12 = *(int *)pcVar12 + 1;
    }
  }
  *(code **)(lVar13 + 0x28) = pcVar11;
  if (((plVar9 != (longlong *)0x0) && (-1 < (int)*plVar9)) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)
     ) {
    (**(code **)(plVar9[1] + 0x30))();
  }
  if (((pcVar12 != (code *)0x0) && (-1 < *(int *)pcVar12)) &&
     (*(longlong *)pcVar12 = *(longlong *)pcVar12 + -1, *(longlong *)pcVar12 == 0)) {
    (**(code **)(*(longlong *)(pcVar12 + 8) + 0x30))(pcVar12);
  }
LAB_1800b6a6a:
  FUN_18036d0a0(plVar7,&DAT_180391c1c,plVar1,plVar2,pcVar3,plVar4);
  plVar9 = DAT_1803b1ea8;
  if (plVar7 == DAT_1803b1ea8) {
    if ((-1 < (int)*DAT_1803b1ea8) && (*DAT_1803b1ea8 = *DAT_1803b1ea8 + -1, *plVar9 == 0)) {
      (**(code **)(plVar9[1] + 0x30))();
    }
    DAT_1803b1ea8 = (longlong *)0x0;
  }
  lVar6 = **(longlong **)(param_1 + 0x38);
  plVar7 = *(longlong **)(lVar6 + 0x30);
  **(longlong **)(param_1 + 0x38) = *(longlong *)(lVar6 + 8);
  *(undefined4 *)(plVar7 + 8) = 0xffffffff;
  plVar9 = (longlong *)plVar7[2];
  if (((plVar9 != (longlong *)0x0) && (plVar7[2] = 0, -1 < (int)*plVar9)) &&
     (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
    (**(code **)(plVar9[1] + 0x30))();
  }
  if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
    (**(code **)(plVar7[1] + 0x30))(plVar7);
  }
  plVar7[10] = 0;
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  if ((-1 < (int)*plVar2) && (*plVar2 = *plVar2 + -1, *plVar2 == 0)) {
    (**(code **)(plVar2[1] + 0x30))(plVar2);
  }
  if ((-1 < *(int *)pcVar3) &&
     (*(longlong *)pcVar3 = *(longlong *)pcVar3 + -1, *(longlong *)pcVar3 == 0)) {
    (**(code **)(*(longlong *)(pcVar3 + 8) + 0x30))(pcVar3);
  }
  if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
    (**(code **)(plVar4[1] + 0x30))(plVar4);
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong *)(param_1 + 0x60) = lVar13;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return (code *)0x0;
}

