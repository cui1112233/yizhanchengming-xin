/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x180160790 */

int * FUN_180160790(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  undefined8 *puVar1;
  int *piVar2;
  longlong *plVar3;
  int iVar4;
  longlong *plVar5;
  longlong *plVar6;
  longlong *plVar7;
  longlong *plVar8;
  longlong *plVar9;
  longlong *plVar10;
  longlong *plVar11;
  longlong lVar12;
  code *pcVar13;
  longlong *plVar14;
  code *pcVar15;
  undefined8 uVar16;
  longlong local_res18;
  longlong *local_res20;
  longlong *local_70;
  longlong *local_68;
  longlong *local_60;
  longlong *local_58;
  longlong *local_50;
  longlong *local_48;

  plVar5 = (longlong *)FUN_18036d7e0(*param_3);
  plVar6 = (longlong *)FUN_18036d7e0(param_3[1]);
  plVar7 = (longlong *)FUN_18036d7e0(param_3[2]);
  plVar8 = (longlong *)FUN_18036d7e0(param_3[3]);
  plVar9 = (longlong *)FUN_18036d7e0(param_3[4]);
  plVar3 = DAT_1803b7ef8;
  local_res20 = DAT_180391d28;
  if (DAT_1803b7ef8 == (longlong *)0x0) {
LAB_180160838:
    DAT_1803b7ef8 = (longlong *)FUN_18036cd10(param_1,DAT_1803b7300,DAT_1803ca750,0x28);
  }
  else if ((1 < *DAT_1803b7ef8) || (DAT_1803b7ef8[2] != 0)) {
    if ((DAT_1803b7ef8 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b7ef8 && (*DAT_1803b7ef8 = *DAT_1803b7ef8 + -1, *plVar3 == 0)))) {
      (**(code **)(plVar3[1] + 0x30))(plVar3);
    }
    goto LAB_180160838;
  }
  plVar3 = DAT_1803b7ef8;
  FUN_180001d00(param_1,DAT_1803b7ef8);
  lVar12 = *(longlong *)(DAT_1803ca748 + 0x20);
  if (*(char *)(lVar12 + 10) == '\0') {
/* novels.pipeline.short_fiction_runner#16="retry_short_fiction_call" */
    plVar10 = (longlong *)FUN_1800015a0(DAT_1803ca748,DAT_1803b7580);
    local_res18 = 0;
    if (plVar10 != (longlong *)0x0) {
      local_res18 = *plVar10;
    }
LAB_18016094b:
    if (local_res18 == 0) goto LAB_180160958;
LAB_1801609a9:
    iVar4 = (int)*plVar6 + 1;
    if (iVar4 != 0) {
      *(int *)plVar6 = iVar4;
    }
    iVar4 = (int)*plVar9 + 1;
    if (iVar4 != 0) {
      *(int *)plVar9 = iVar4;
    }
    iVar4 = (int)*plVar7 + 1;
    if (iVar4 != 0) {
      *(int *)plVar7 = iVar4;
    }
    iVar4 = (int)*plVar5 + 1;
    if (iVar4 != 0) {
      *(int *)plVar5 = iVar4;
    }
    iVar4 = (int)*plVar8 + 1;
    if (iVar4 != 0) {
      *(int *)plVar8 = iVar4;
    }
    local_68 = plVar6;
    local_60 = plVar9;
    local_58 = plVar7;
    local_50 = plVar5;
    local_48 = plVar8;
    plVar11 = (longlong *)
/* novels.pipeline.short_fiction_runner#2="<lambda>" ; novels.pipeline.short_fiction_runner#17="ShortFictionRunner._chat.<locals>.<lambda>" */
              FUN_1803609a0(FUN_180160f00,DAT_1803b7510,DAT_1803b7588,DAT_1803b73a0,0,0,0,
                            DAT_1803ca750,0,&local_68,5);
    lVar12 = *(longlong *)(param_1 + 0x10);
/* novels.pipeline.short_fiction_runner#18="short-fiction-" */
    local_res20 = DAT_1803b7590;
    plVar10 = *(longlong **)(lVar12 + 0x41da0);
    if (plVar10 == (longlong *)0x0) {
      plVar10 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,2);
    }
    else {
      *(longlong *)(lVar12 + 0x41da0) = plVar10[3];
      *(int *)(lVar12 + 0x41e3c) = *(int *)(lVar12 + 0x41e3c) + -1;
      *plVar10 = 1;
    }
    plVar10[3] = 0;
    plVar10[4] = 0;
    plVar14 = plVar10 + -2;
    lVar12 = _PyThreadState_GetCurrent();
    lVar12 = *(longlong *)(*(longlong *)(lVar12 + 0x10) + 0xd0);
    puVar1 = *(undefined8 **)(lVar12 + 8);
    *puVar1 = plVar14;
    plVar10[-1] = plVar10[-1] & 3;
    plVar10[-1] = plVar10[-1] | (ulonglong)puVar1;
    *plVar14 = lVar12;
    *(longlong **)(lVar12 + 8) = plVar14;
    iVar4 = (int)*local_res20 + 1;
    if (iVar4 != 0) {
      *(int *)local_res20 = iVar4;
    }
    plVar10[3] = (longlong)local_res20;
/* novels.pipeline.short_fiction_runner#19="" */
    lVar12 = FUN_180375520(param_1,plVar6[2],DAT_1803b7598);
    if (lVar12 == 0) {
      plVar14 = *(longlong **)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
        (**(code **)(plVar11[1] + 0x30))(plVar11);
      }
      if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
        pcVar13 = *(code **)(plVar10[1] + 0x30);
        plVar11 = plVar10;
LAB_180160b57:
        (*pcVar13)(plVar11);
      }
    }
    else {
      plVar10[4] = lVar12;
/* novels.pipeline.short_fiction_runner#19="" */
      plVar14 = (longlong *)PyUnicode_Join(DAT_1803b7598,plVar10);
      if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
        (**(code **)(plVar10[1] + 0x30))(plVar10);
      }
      if (plVar14 != (longlong *)0x0) {
        uVar16 = 0x69;
        *(undefined4 *)(plVar3 + 5) = 0x69;
        local_res20 = plVar14;
        local_70 = plVar11;
        local_res20 = (longlong *)
/* novels.pipeline.short_fiction_runner#20={"type": "tuple", "items": [{"tag": "a", "fileRange": [4003604, 4003611], "value": "label"}]} */
                      FUN_18037d030(param_1,local_res18,&local_70,&local_res20,DAT_1803b75a0);
        if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
          (**(code **)(plVar11[1] + 0x30))(plVar11);
        }
        if ((-1 < (int)*plVar14) && (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
          (**(code **)(plVar14[1] + 0x30))(plVar14);
        }
        plVar10 = local_res20;
        if (local_res20 != (longlong *)0x0) {
          FUN_180001d80(param_1);
          if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
            (**(code **)(plVar5[1] + 0x30))();
          }
          if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
            (**(code **)(plVar6[1] + 0x30))(plVar6);
          }
          if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
            (**(code **)(plVar7[1] + 0x30))(plVar7);
          }
          if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
            (**(code **)(plVar8[1] + 0x30))(plVar8);
          }
          if ((int)*plVar9 < 0) {
            return (int *)plVar10;
          }
          *plVar9 = *plVar9 + -1;
          if (*plVar9 != 0) {
            return (int *)plVar10;
          }
          (**(code **)(plVar9[1] + 0x30))(plVar9);
          return (int *)plVar10;
        }
        plVar14 = *(longlong **)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
        goto LAB_180160b5e;
      }
      plVar14 = *(longlong **)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
        pcVar13 = *(code **)(plVar11[1] + 0x30);
        goto LAB_180160b57;
      }
    }
    uVar16 = 0x6b;
  }
  else {
    iVar4 = *(int *)(lVar12 + 0xc);
    if (iVar4 == 0) {
      iVar4 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
      *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar4 + 1;
      *(int *)(lVar12 + 0xc) = iVar4;
    }
    if (iVar4 != DAT_1803a9980) {
      DAT_1803a9980 = iVar4;
/* novels.pipeline.short_fiction_runner#16="retry_short_fiction_call" */
      DAT_1803b7318 = FUN_1803707d0(lVar12,DAT_1803b7580,*(undefined8 *)(DAT_1803b7580 + 0x18));
    }
    if (-1 < DAT_1803b7318) {
      local_70 = (longlong *)((1L << (*(byte *)(lVar12 + 9) & 0x3f)) + 0x20 + lVar12);
      local_res18 = local_70[DAT_1803b7318 * 2 + 1];
      if (local_res18 != 0) goto LAB_1801609a9;
/* novels.pipeline.short_fiction_runner#16="retry_short_fiction_call" */
      DAT_1803b7318 = FUN_1803707d0(lVar12,DAT_1803b7580,*(undefined8 *)(DAT_1803b7580 + 0x18));
      if (-1 < DAT_1803b7318) {
        local_res18 = local_70[DAT_1803b7318 * 2 + 1];
        goto LAB_18016094b;
      }
    }
LAB_180160958:
/* novels.pipeline.short_fiction_runner#16="retry_short_fiction_call" */
    plVar10 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803b7580);
    local_res18 = 0;
    if (plVar10 != (longlong *)0x0) {
      local_res18 = *plVar10;
    }
    if (local_res18 != 0) goto LAB_1801609a9;
/* novels.pipeline.short_fiction_runner#16="retry_short_fiction_call" */
    FUN_180373340(param_1,&local_res20,DAT_1803b7580);
    uVar16 = 0x69;
    plVar14 = local_res20;
  }
LAB_180160b5e:
  piVar2 = (int *)plVar14[5];
  if (piVar2 == (int *)0x0) {
    pcVar13 = (code *)FUN_18038e2a0(plVar3,uVar16);
    plVar10 = (longlong *)plVar14[5];
    if (pcVar13 == _Py_NoneStruct_exref) {
      pcVar15 = (code *)0x0;
    }
    else {
      pcVar15 = pcVar13;
      if ((pcVar13 != (code *)0x0) && (*(int *)pcVar13 + 1 != 0)) {
        *(int *)pcVar13 = *(int *)pcVar13 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar2 + 6) == plVar3) goto LAB_180160d3e;
    pcVar13 = (code *)FUN_18038e2a0(plVar3,uVar16);
    *(int **)(pcVar13 + 0x10) = piVar2;
    if (*piVar2 + 1 != 0) {
      *piVar2 = *piVar2 + 1;
    }
    plVar10 = (longlong *)plVar14[5];
    if (pcVar13 == _Py_NoneStruct_exref) {
      pcVar15 = (code *)0x0;
    }
    else {
      pcVar15 = pcVar13;
      if ((pcVar13 != (code *)0x0) && (*(int *)pcVar13 + 1 != 0)) {
        *(int *)pcVar13 = *(int *)pcVar13 + 1;
      }
    }
  }
  plVar14[5] = (longlong)pcVar15;
  if (((plVar10 != (longlong *)0x0) && (-1 < (int)*plVar10)) &&
     (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
    (**(code **)(plVar10[1] + 0x30))();
  }
  if (((pcVar13 != (code *)0x0) && (-1 < *(int *)pcVar13)) &&
     (*(longlong *)pcVar13 = *(longlong *)pcVar13 + -1, *(longlong *)pcVar13 == 0)) {
    (**(code **)(*(longlong *)(pcVar13 + 8) + 0x30))(pcVar13);
  }
LAB_180160d3e:
  FUN_18036d0a0(plVar3,"ccccc",plVar5,plVar6,plVar7,plVar8,plVar9);
  plVar10 = DAT_1803b7ef8;
  if (plVar3 == DAT_1803b7ef8) {
    if ((-1 < (int)*DAT_1803b7ef8) && (*DAT_1803b7ef8 = *DAT_1803b7ef8 + -1, *plVar10 == 0)) {
      (**(code **)(plVar10[1] + 0x30))();
    }
    DAT_1803b7ef8 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
    (**(code **)(plVar5[1] + 0x30))(plVar5);
  }
  if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
    (**(code **)(plVar6[1] + 0x30))(plVar6);
  }
  if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
    (**(code **)(plVar7[1] + 0x30))(plVar7);
  }
  if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
    (**(code **)(plVar8[1] + 0x30))(plVar8);
  }
  if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
    (**(code **)(plVar9[1] + 0x30))(plVar9);
  }
  plVar3 = *(longlong **)(param_1 + 0x60);
  *(longlong **)(param_1 + 0x60) = plVar14;
  if (((plVar3 != (longlong *)0x0) && (-1 < (int)*plVar3)) && (*plVar3 = *plVar3 + -1, *plVar3 == 0)
     ) {
    (**(code **)(plVar3[1] + 0x30))();
  }
  return (int *)0x0;
}
