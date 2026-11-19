/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x1801c5cd0 */

longlong FUN_1801c5cd0(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  undefined8 *puVar2;
  int *piVar3;
  longlong *plVar4;
  int iVar5;
  longlong *plVar6;
  longlong *plVar7;
  longlong *plVar8;
  longlong lVar9;
  longlong lVar10;
  longlong *plVar11;
  code *pcVar12;
  code *pcVar13;
  longlong lVar14;
  undefined8 uVar15;
  longlong local_res20;
  undefined8 local_48;
  longlong *local_40;
  
  plVar1 = (longlong *)*param_3;
  plVar6 = (longlong *)FUN_18036d6d0();
  plVar4 = DAT_1803bc490;
  local_res20 = DAT_180391dc8;
  if (DAT_1803bc490 == (longlong *)0x0) {
LAB_1801c5d30:
    DAT_1803bc490 = (longlong *)FUN_18036cd10(param_1,DAT_1803bbdf0,DAT_1803ca610,0x10);
  }
  else if ((1 < *DAT_1803bc490) || (DAT_1803bc490[2] != 0)) {
    if ((DAT_1803bc490 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803bc490 && (*DAT_1803bc490 = *DAT_1803bc490 + -1, *plVar4 == 0)))) {
      (**(code **)(plVar4[1] + 0x30))(plVar4);
    }
    goto LAB_1801c5d30;
  }
  plVar4 = DAT_1803bc490;
  FUN_180001d00(param_1,DAT_1803bc490);
  plVar7 = (longlong *)FUN_18037f0c0();
  uVar15 = 0x79;
  if (plVar7 == (longlong *)0x0) {
LAB_1801c6194:
    lVar9 = *(longlong *)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
  }
  else {
    *(undefined4 *)(plVar4 + 5) = 0x79;
    plVar8 = (longlong *)FUN_180378210(param_1,plVar7,DAT_1803bbe20);
    if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
      (**(code **)(plVar7[1] + 0x30))(plVar7);
    }
    if (plVar8 == (longlong *)0x0) goto LAB_1801c6194;
/* novels.prompts.inkos.reviser_prompts#3="" */
    lVar9 = FUN_180375520(param_1,plVar8,DAT_1803bbe28);
    if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
      (**(code **)(plVar8[1] + 0x30))(plVar8);
    }
    if (lVar9 == 0) goto LAB_1801c6194;
    lVar10 = *(longlong *)(param_1 + 0x10);
    plVar7 = *(longlong **)(lVar10 + 0x41da8);
    if (plVar7 == (longlong *)0x0) {
      plVar7 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,3);
    }
    else {
      *(longlong *)(lVar10 + 0x41da8) = plVar7[3];
      *(int *)(lVar10 + 0x41e40) = *(int *)(lVar10 + 0x41e40) + -1;
      *plVar7 = 1;
    }
    lVar14 = 0;
    plVar8 = plVar7 + -2;
    plVar7[3] = 0;
    plVar7[4] = 0;
    plVar7[5] = 0;
    lVar10 = _PyThreadState_GetCurrent();
    lVar10 = *(longlong *)(*(longlong *)(lVar10 + 0x10) + 0xd0);
    puVar2 = *(undefined8 **)(lVar10 + 8);
    *puVar2 = plVar8;
    plVar7[-1] = plVar7[-1] & 3;
    plVar7[-1] = plVar7[-1] | (ulonglong)puVar2;
    *plVar8 = lVar10;
    *(longlong **)(lVar10 + 8) = plVar8;
    plVar7[3] = lVar9;
/* novels.prompts.inkos.reviser_prompts#35=" " */
    piVar3 = DAT_1803bbf28;
/* novels.prompts.inkos.reviser_prompts#35=" " */
    if (*DAT_1803bbf28 + 1 != 0) {
/* novels.prompts.inkos.reviser_prompts#35=" " */
      *DAT_1803bbf28 = *DAT_1803bbf28 + 1;
    }
    plVar7[4] = (longlong)piVar3;
/* novels.prompts.inkos.reviser_prompts#1="get" */
    plVar8 = (longlong *)FUN_18037f0c0(piVar3,plVar1,DAT_1803bbe18);
    if (plVar8 == (longlong *)0x0) {
LAB_1801c5f4f:
      lVar9 = *(longlong *)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
        (**(code **)(plVar7[1] + 0x30))(plVar7);
      }
    }
    else {
      *(undefined4 *)(plVar4 + 5) = 0x79;
      plVar11 = (longlong *)FUN_180378210(param_1,plVar8,DAT_1803bbe38);
      if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
        (**(code **)(plVar8[1] + 0x30))(plVar8);
      }
      if (plVar11 == (longlong *)0x0) goto LAB_1801c5f4f;
/* novels.prompts.inkos.reviser_prompts#3="" */
      lVar9 = FUN_180375520(param_1,plVar11,DAT_1803bbe28);
      if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
        (**(code **)(plVar11[1] + 0x30))(plVar11);
      }
      if (lVar9 == 0) goto LAB_1801c5f4f;
      plVar7[5] = lVar9;
/* novels.prompts.inkos.reviser_prompts#3="" */
      lVar9 = PyUnicode_Join(DAT_1803bbe28,plVar7);
      if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
        (**(code **)(plVar7[1] + 0x30))(plVar7);
      }
      lVar10 = DAT_1803ca608;
      if (lVar9 == 0) goto LAB_1801c6194;
      plVar6[2] = lVar9;
      lVar9 = *(longlong *)(lVar10 + 0x20);
      if (*(char *)(lVar9 + 10) == '\0') {
/* novels.prompts.inkos.reviser_prompts#40="LOCAL_ONLY_PATTERNS" */
        plVar7 = (longlong *)FUN_1800015a0(lVar10,DAT_1803bbf50);
        if (plVar7 != (longlong *)0x0) {
          lVar14 = *plVar7;
        }
LAB_1801c6080:
        if (lVar14 == 0) goto LAB_1801c6085;
LAB_1801c60ce:
        plVar7 = (longlong *)FUN_18000d280();
        if (plVar7 != (longlong *)0x0) {
          local_48 = FUN_18036d750(plVar7);
          iVar5 = (int)*plVar6 + 1;
          if (iVar5 != 0) {
            *(int *)plVar6 = iVar5;
          }
          local_40 = plVar6;
          plVar8 = (longlong *)
/* novels.prompts.inkos.reviser_prompts#32="<genexpr>" ; novels.prompts.inkos.reviser_prompts#41="resolve_auto_output_mode.<locals>.is_local.<locals>.<genexpr>" */
                   FUN_180367890(FUN_1801c6340,DAT_1803ca610,DAT_1803bbf10,DAT_1803bbf58,
                                 DAT_1803bbda8,&local_48,2,0x448);
          if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
            (**(code **)(plVar7[1] + 0x30))(plVar7);
          }
          lVar9 = FUN_1803753b0(param_1,plVar8);
          if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
            (**(code **)(plVar8[1] + 0x30))(plVar8);
          }
          if (lVar9 != 0) {
            FUN_180001d80(param_1);
            if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
              (**(code **)(plVar6[1] + 0x30))(plVar6);
            }
            if ((int)*plVar1 < 0) {
              return lVar9;
            }
            *plVar1 = *plVar1 + -1;
            if (*plVar1 != 0) {
              return lVar9;
            }
            (**(code **)(plVar1[1] + 0x30))(plVar1);
            return lVar9;
          }
        }
        uVar15 = 0x7a;
        goto LAB_1801c6194;
      }
      iVar5 = *(int *)(lVar9 + 0xc);
      if (iVar5 == 0) {
        iVar5 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
        *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar5 + 1;
        *(int *)(lVar9 + 0xc) = iVar5;
      }
      if (iVar5 != DAT_1803a9ad8) {
        DAT_1803a9ad8 = iVar5;
/* novels.prompts.inkos.reviser_prompts#40="LOCAL_ONLY_PATTERNS" */
        DAT_1803bbdd8 = FUN_1803707d0(lVar9,DAT_1803bbf50,*(undefined8 *)(DAT_1803bbf50 + 0x18));
      }
      if (-1 < DAT_1803bbdd8) {
        lVar10 = (1L << (*(byte *)(lVar9 + 9) & 0x3f)) + lVar9;
        if (*(longlong *)(lVar10 + 0x28 + DAT_1803bbdd8 * 0x10) != 0) goto LAB_1801c60ce;
/* novels.prompts.inkos.reviser_prompts#40="LOCAL_ONLY_PATTERNS" */
        DAT_1803bbdd8 = FUN_1803707d0(lVar9,DAT_1803bbf50,*(undefined8 *)(DAT_1803bbf50 + 0x18));
        if (-1 < DAT_1803bbdd8) {
          lVar14 = *(longlong *)(lVar10 + 0x28 + DAT_1803bbdd8 * 0x10);
          goto LAB_1801c6080;
        }
      }
LAB_1801c6085:
/* novels.prompts.inkos.reviser_prompts#40="LOCAL_ONLY_PATTERNS" */
      plVar7 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803bbf50);
      if ((plVar7 != (longlong *)0x0) && (*plVar7 != 0)) goto LAB_1801c60ce;
/* novels.prompts.inkos.reviser_prompts#40="LOCAL_ONLY_PATTERNS" */
      FUN_180373340(param_1,&local_res20,DAT_1803bbf50);
      uVar15 = 0x7a;
      lVar9 = local_res20;
    }
  }
  pcVar13 = (code *)0x0;
  piVar3 = *(int **)(lVar9 + 0x28);
  if (piVar3 == (int *)0x0) {
    pcVar12 = (code *)FUN_18038e2a0(plVar4,uVar15);
    plVar7 = *(longlong **)(lVar9 + 0x28);
    if (pcVar12 != _Py_NoneStruct_exref) {
LAB_1801c61f8:
      pcVar13 = pcVar12;
      pcVar12 = pcVar13;
      if ((pcVar13 != (code *)0x0) && (*(int *)pcVar13 + 1 != 0)) {
        *(int *)pcVar13 = *(int *)pcVar13 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar3 + 6) == plVar4) goto LAB_1801c6241;
    pcVar12 = (code *)FUN_18038e2a0(plVar4,uVar15);
    *(int **)(pcVar12 + 0x10) = piVar3;
    if (*piVar3 + 1 != 0) {
      *piVar3 = *piVar3 + 1;
    }
    plVar7 = *(longlong **)(lVar9 + 0x28);
    if (pcVar12 != _Py_NoneStruct_exref) goto LAB_1801c61f8;
  }
  *(code **)(lVar9 + 0x28) = pcVar13;
  if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)
     ) {
    (**(code **)(plVar7[1] + 0x30))();
  }
  if (((pcVar12 != (code *)0x0) && (-1 < *(int *)pcVar12)) &&
     (*(longlong *)pcVar12 = *(longlong *)pcVar12 + -1, *(longlong *)pcVar12 == 0)) {
    (**(code **)(*(longlong *)(pcVar12 + 8) + 0x30))(pcVar12);
  }
LAB_1801c6241:
  FUN_18036d0a0(plVar4,&DAT_18039495c,plVar1,plVar6);
  plVar7 = DAT_1803bc490;
  if (plVar4 == DAT_1803bc490) {
    if ((-1 < (int)*DAT_1803bc490) && (*DAT_1803bc490 = *DAT_1803bc490 + -1, *plVar7 == 0)) {
      (**(code **)(plVar7[1] + 0x30))();
    }
    DAT_1803bc490 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
    (**(code **)(plVar6[1] + 0x30))(plVar6);
  }
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))(plVar1);
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong *)(param_1 + 0x60) = lVar9;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return 0;
}
