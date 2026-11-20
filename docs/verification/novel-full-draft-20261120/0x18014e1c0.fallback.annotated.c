/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x18014e1c0 */

int * FUN_18014e1c0(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  longlong *plVar2;
  longlong *plVar3;
  int iVar4;
  longlong lVar5;
  longlong *plVar6;
  int *piVar7;
  code *pcVar8;
  code *pcVar9;
  longlong *plVar10;
  undefined8 uVar11;
  longlong local_res18 [2];
  longlong *local_48;
  longlong *local_40;

  plVar3 = DAT_1803b6b18;
  plVar10 = (longlong *)0x0;
  plVar1 = (longlong *)*param_3;
  plVar2 = (longlong *)param_3[1];
  local_res18[0] = DAT_180391d20;
  if (DAT_1803b6b18 == (longlong *)0x0) {
LAB_18014e222:
    DAT_1803b6b18 = (longlong *)FUN_18036cd10(param_1,DAT_1803b6b48,DAT_1803ca758,0x18);
  }
  else if ((1 < *DAT_1803b6b18) || (DAT_1803b6b18[2] != 0)) {
    if ((DAT_1803b6b18 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b6b18 && (*DAT_1803b6b18 = *DAT_1803b6b18 + -1, *plVar3 == 0)))) {
      (**(code **)(plVar3[1] + 0x30))(plVar3);
    }
    goto LAB_18014e222;
  }
  plVar3 = DAT_1803b6b18;
  FUN_180001d00(param_1,DAT_1803b6b18);
  lVar5 = FUN_18014c130(param_1);
  if (lVar5 == 0) {
/* novels.pipeline.short_fiction_parsers#18="_extract_tagged_blocks" */
    FUN_180373340(param_1,local_res18,DAT_1803b6cc0);
    uVar11 = 0x49;
    lVar5 = local_res18[0];
  }
  else {
    uVar11 = 0x49;
    *(undefined4 *)(plVar3 + 5) = 0x49;
    local_48 = plVar1;
    local_40 = plVar2;
    plVar6 = (longlong *)FUN_180377750(param_1,lVar5,&local_48);
    if (plVar6 != (longlong *)0x0) {
      iVar4 = FUN_180001510(plVar6);
/* novels.pipeline.short_fiction_parsers#4="" */
      piVar7 = DAT_1803b6c50;
      if (iVar4 != -1) {
        if (iVar4 == 0) {
/* novels.pipeline.short_fiction_parsers#4="" */
          if (*DAT_1803b6c50 + 1 != 0) {
/* novels.pipeline.short_fiction_parsers#4="" */
            *DAT_1803b6c50 = *DAT_1803b6c50 + 1;
          }
LAB_18014e484:
          FUN_180001d80(param_1);
          if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
            (**(code **)(plVar6[1] + 0x30))(plVar6);
          }
          if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
            (**(code **)(plVar1[1] + 0x30))(plVar1);
          }
          if ((int)*plVar2 < 0) {
            return piVar7;
          }
          *plVar2 = *plVar2 + -1;
          if (*plVar2 != 0) {
            return piVar7;
          }
          (**(code **)(plVar2[1] + 0x30))(plVar2);
          return piVar7;
        }
        piVar7 = (int *)FUN_18014b9b0(param_1,plVar6,DAT_1803ab250,0);
        if (piVar7 != (int *)0x0) goto LAB_18014e484;
      }
      uVar11 = 0x4a;
      plVar10 = plVar6;
    }
    lVar5 = *(longlong *)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
  }
  piVar7 = *(int **)(lVar5 + 0x28);
  if (piVar7 == (int *)0x0) {
    pcVar8 = (code *)FUN_18038e2a0(plVar3,uVar11);
    plVar6 = *(longlong **)(lVar5 + 0x28);
    pcVar9 = pcVar8;
    if (pcVar8 == _Py_NoneStruct_exref) {
      pcVar9 = (code *)0x0;
    }
    else {
LAB_18014e362:
      pcVar8 = pcVar9;
      if ((pcVar9 != (code *)0x0) && (*(int *)pcVar9 + 1 != 0)) {
        *(int *)pcVar9 = *(int *)pcVar9 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar7 + 6) == plVar3) goto LAB_18014e3ac;
    pcVar8 = (code *)FUN_18038e2a0(plVar3,uVar11);
    *(int **)(pcVar8 + 0x10) = piVar7;
    if (*piVar7 + 1 != 0) {
      *piVar7 = *piVar7 + 1;
    }
    plVar6 = *(longlong **)(lVar5 + 0x28);
    pcVar9 = pcVar8;
    if (pcVar8 != _Py_NoneStruct_exref) goto LAB_18014e362;
    pcVar9 = (code *)0x0;
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
LAB_18014e3ac:
  FUN_18036d0a0(plVar3,&DAT_1803948e8,plVar1,plVar2,plVar10);
  plVar6 = DAT_1803b6b18;
  if (plVar3 == DAT_1803b6b18) {
    if ((-1 < (int)*DAT_1803b6b18) && (*DAT_1803b6b18 = *DAT_1803b6b18 + -1, *plVar6 == 0)) {
      (**(code **)(plVar6[1] + 0x30))();
    }
    DAT_1803b6b18 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if (((plVar10 != (longlong *)0x0) && (-1 < (int)*plVar10)) &&
     (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
    (**(code **)(plVar10[1] + 0x30))(plVar10);
  }
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))(plVar1);
  }
  if ((-1 < (int)*plVar2) && (*plVar2 = *plVar2 + -1, *plVar2 == 0)) {
    (**(code **)(plVar2[1] + 0x30))(plVar2);
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong *)(param_1 + 0x60) = lVar5;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return (int *)0x0;
}
