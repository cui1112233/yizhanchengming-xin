longlong * FUN_1800d31d0(longlong param_1,undefined8 param_2,undefined8 *param_3)
{
  longlong *plVar1;
  longlong *plVar2;
  undefined8 *puVar3;
  int *piVar4;
  longlong *plVar5;
  undefined8 uVar6;
  int iVar7;
  longlong *plVar8;
  longlong lVar9;
  longlong *plVar10;
  longlong *plVar11;
  longlong *plVar12;
  code *pcVar13;
  code *pcVar14;
  int iVar15;
  undefined8 uVar16;
  longlong lVar17;
  longlong lVar18;
  longlong *local_res20;
  longlong *local_68;
  longlong *local_60;
  longlong *local_58;
  longlong *local_50;
  longlong *local_48;
  plVar5 = DAT_1803b2a18;
  local_50 = (longlong *)*param_3;
  plVar1 = (longlong *)param_3[1];
  plVar2 = (longlong *)param_3[2];
  local_res20 = (longlong *)0x0;
  local_68 = (longlong *)0x0;
  if (DAT_1803b2a18 == (longlong *)0x0) {
LAB_1800d3246:
    DAT_1803b2a18 = (longlong *)FUN_18036cd10(param_1,DAT_1803b2ac0,DAT_1803ca7a0,0x20);
  }
  else if ((1 < *DAT_1803b2a18) || (DAT_1803b2a18[2] != 0)) {
    if ((DAT_1803b2a18 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b2a18 && (*DAT_1803b2a18 = *DAT_1803b2a18 + -1, *plVar5 == 0)))) {
      (**(code **)(plVar5[1] + 0x30))(plVar5);
    }
    goto LAB_1800d3246;
  }
  plVar5 = DAT_1803b2a18;
  FUN_180001d00(param_1,DAT_1803b2a18);
  plVar8 = (longlong *)FUN_18037f0c0();
/* Analyst constant binding: novels.pipeline.imitation_short_runner#176="auditor" bytes=[3955744, 3955753] */
  piVar4 = DAT_1803b3080;
  if (plVar8 == (longlong *)0x0) {
    plVar11 = *(longlong **)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
    uVar16 = 0x1a1;
    goto LAB_1800d3929;
  }
  lVar9 = *(longlong *)(param_1 + 0x10);
  plVar12 = *(longlong **)(lVar9 + 0x41da0);
  if (plVar12 == (longlong *)0x0) {
    plVar12 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,2);
  }
  else {
    *(longlong *)(lVar9 + 0x41da0) = plVar12[3];
    *(int *)(lVar9 + 0x41e3c) = *(int *)(lVar9 + 0x41e3c) + -1;
    *plVar12 = 1;
  }
  plVar11 = plVar12 + -2;
  plVar12[3] = 0;
  plVar12[4] = 0;
  lVar9 = _PyThreadState_GetCurrent();
  lVar9 = *(longlong *)(*(longlong *)(lVar9 + 0x10) + 0xd0);
  puVar3 = *(undefined8 **)(lVar9 + 8);
  *puVar3 = plVar11;
  plVar12[-1] = plVar12[-1] & 3;
  plVar12[-1] = plVar12[-1] | (ulonglong)puVar3;
  *plVar11 = lVar9;
  *(longlong **)(lVar9 + 8) = plVar11;
  if (*piVar4 + 1 != 0) {
    *piVar4 = *piVar4 + 1;
  }
  plVar12[3] = (longlong)piVar4;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#153="system" bytes=[3954951, 3954959] */
  uVar6 = DAT_1803b2fc8;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#152="role" bytes=[3954945, 3954951] */
  uVar16 = DAT_1803b2fc0;
  plVar10 = (longlong *)_PyDict_NewPresized(2);
  PyDict_SetItem(plVar10,uVar16,uVar6);
/* Analyst constant binding: novels.pipeline.imitation_short_runner#154="content" bytes=[3954959, 3954968] */
  local_58 = DAT_1803b2fd0;
  lVar9 = *(longlong *)(DAT_1803ca798 + 0x20);
  if (*(char *)(lVar9 + 10) == '\0') {
    plVar11 = (longlong *)FUN_1800015a0();
    if (plVar11 == (longlong *)0x0) {
      lVar9 = 0;
    }
    else {
      lVar9 = *plVar11;
    }
LAB_1800d3433:
    if (lVar9 == 0) goto LAB_1800d3438;
LAB_1800d3478:
    *(undefined4 *)(plVar5 + 5) = 0x1a4;
    plVar11 = (longlong *)FUN_180376210(param_1);
    if (plVar11 == (longlong *)0x0) {
      plVar11 = *(longlong **)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      goto LAB_1800d3499;
    }
    PyDict_SetItem(plVar10,local_58,plVar11);
    if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
      (**(code **)(plVar11[1] + 0x30))(plVar11);
    }
    local_58 = (longlong *)FUN_180372160(param_1,2);
    *(longlong **)local_58[3] = plVar10;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#156="user" bytes=[3955004, 3955010] */
    uVar6 = DAT_1803b2fe0;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#152="role" bytes=[3954945, 3954951] */
    uVar16 = DAT_1803b2fc0;
    plVar10 = (longlong *)_PyDict_NewPresized(2);
    PyDict_SetItem(plVar10,uVar16,uVar6);
/* Analyst constant binding: novels.pipeline.imitation_short_runner#154="content" bytes=[3954959, 3954968] */
    local_48 = DAT_1803b2fd0;
    lVar9 = *(longlong *)(DAT_1803ca798 + 0x20);
    if (*(char *)(lVar9 + 10) != '\0') {
      iVar15 = *(int *)(lVar9 + 0xc);
      if (iVar15 == 0) {
        iVar15 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
        *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar15 + 1;
        *(int *)(lVar9 + 0xc) = iVar15;
      }
      if (iVar15 != DAT_1803a977c) {
        DAT_1803a977c = iVar15;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#178="build_editorial_review_user_prompt" bytes=[3955791, 3955827] */
        DAT_1803b2980 = FUN_1803707d0(lVar9,DAT_1803b3090,*(undefined8 *)(DAT_1803b3090 + 0x18));
      }
      if (-1 < DAT_1803b2980) {
        lVar18 = (1L << (*(byte *)(lVar9 + 9) & 0x3f)) + lVar9;
        lVar17 = *(longlong *)(lVar18 + 0x28 + DAT_1803b2980 * 0x10);
        if (lVar17 != 0) goto LAB_1800d3643;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#178="build_editorial_review_user_prompt" bytes=[3955791, 3955827] */
        DAT_1803b2980 = FUN_1803707d0(lVar9,DAT_1803b3090,*(undefined8 *)(DAT_1803b3090 + 0x18));
        if (-1 < DAT_1803b2980) {
          lVar17 = *(longlong *)(lVar18 + 0x28 + DAT_1803b2980 * 0x10);
          goto LAB_1800d35fe;
        }
      }
LAB_1800d3603:
/* Analyst constant binding: novels.pipeline.imitation_short_runner#178="build_editorial_review_user_prompt" bytes=[3955791, 3955827] */
      plVar11 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803b3090);
      if ((plVar11 != (longlong *)0x0) && (lVar17 = *plVar11, lVar17 != 0)) goto LAB_1800d3643;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#178="build_editorial_review_user_prompt" bytes=[3955791, 3955827] */
      FUN_180373340(param_1,&local_68,DAT_1803b3090);
      plVar11 = local_68;
LAB_1800d366c:
      uVar16 = 0x1a5;
      if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
        (**(code **)(plVar10[1] + 0x30))(plVar10);
      }
      if ((-1 < (int)*local_58) && (*local_58 = *local_58 + -1, *local_58 == 0)) {
        lVar9 = local_58[1];
        plVar10 = local_58;
        goto LAB_1800d369c;
      }
      goto LAB_1800d36a2;
    }
    plVar11 = (longlong *)FUN_1800015a0();
    if (plVar11 == (longlong *)0x0) {
      lVar17 = 0;
    }
    else {
      lVar17 = *plVar11;
    }
LAB_1800d35fe:
    if (lVar17 == 0) goto LAB_1800d3603;
LAB_1800d3643:
    *(undefined4 *)(plVar5 + 5) = 0x1a5;
    plVar11 = (longlong *)FUN_180376770(param_1,lVar17,plVar1);
    if (plVar11 == (longlong *)0x0) {
      plVar11 = *(longlong **)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      goto LAB_1800d366c;
    }
    PyDict_SetItem(plVar10,local_48,plVar11);
    if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
      (**(code **)(plVar11[1] + 0x30))(plVar11);
    }
    *(longlong **)(local_58[3] + 8) = plVar10;
    plVar12[4] = (longlong)local_58;
    plVar11 = (longlong *)FUN_18036fb80(param_1,DAT_1803b3098);
    uVar16 = 0x1a1;
    *(undefined4 *)(plVar5 + 5) = 0x1a1;
    plVar10 = (longlong *)FUN_180001790(param_1,plVar8,plVar12,plVar11);
    if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
      (**(code **)(plVar8[1] + 0x30))(plVar8);
    }
    if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
      (**(code **)(plVar12[1] + 0x30))(plVar12);
    }
    if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
      (**(code **)(plVar11[1] + 0x30))(plVar11);
    }
    if (plVar10 == (longlong *)0x0) {
LAB_1800d390d:
      plVar11 = *(longlong **)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
    }
    else {
      lVar9 = FUN_1800c7020(param_1);
      local_res20 = plVar10;
      if (lVar9 == 0) {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#4="FileRepository" bytes=[3952488, 3952504] */
        FUN_180373340(param_1,&local_68,DAT_1803b2b20);
        uVar16 = 0x1ab;
        plVar11 = local_68;
      }
      else {
        plVar8 = (longlong *)FUN_18037f0c0();
        if (plVar8 == (longlong *)0x0) {
          plVar11 = *(longlong **)(param_1 + 0x60);
          uVar16 = 0x1ab;
          *(undefined8 *)(param_1 + 0x60) = 0;
        }
        else {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#52="reviews" bytes=[3953291, 3953300] */
          plVar11 = (longlong *)FUN_180382540(plVar2,DAT_1803b2ca0);
          if (plVar11 == (longlong *)0x0) {
            plVar11 = *(longlong **)(param_1 + 0x60);
            *(undefined8 *)(param_1 + 0x60) = 0;
            if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
              (**(code **)(plVar8[1] + 0x30))(plVar8);
            }
            uVar16 = 0x1ab;
          }
          else {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#144="editorial.md" bytes=[3954798, 3954812] */
            plVar12 = (longlong *)FUN_180382540(plVar11,DAT_1803b2f80);
            if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
              (**(code **)(plVar11[1] + 0x30))(plVar11);
            }
            if (plVar12 != (longlong *)0x0) {
              uVar16 = 0x1ab;
              *(undefined4 *)(plVar5 + 5) = 0x1ab;
              local_68 = plVar12;
              local_60 = plVar10;
              plVar11 = (longlong *)FUN_180377750(param_1,plVar8,&local_68);
              if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                (**(code **)(plVar8[1] + 0x30))(plVar8);
              }
              if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                (**(code **)(plVar12[1] + 0x30))(plVar12);
              }
              if (plVar11 != (longlong *)0x0) {
                if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                  (**(code **)(plVar11[1] + 0x30))(plVar11);
                }
                FUN_180001d80(param_1);
                iVar15 = (int)*plVar10 + 1;
                iVar7 = (int)*plVar10;
                if (iVar15 != 0) {
                  *(int *)plVar10 = iVar15;
                  iVar7 = iVar15;
                }
                if ((-1 < iVar7) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                  (**(code **)(plVar10[1] + 0x30))(plVar10);
                }
                if ((-1 < (int)*local_50) && (*local_50 = *local_50 + -1, *local_50 == 0)) {
                  (**(code **)(local_50[1] + 0x30))();
                }
                if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
                  (**(code **)(plVar1[1] + 0x30))();
                }
                if ((int)*plVar2 < 0) {
                  return plVar10;
                }
                *plVar2 = *plVar2 + -1;
                if (*plVar2 != 0) {
                  return plVar10;
                }
                (**(code **)(plVar2[1] + 0x30))(plVar2);
                return plVar10;
              }
              goto LAB_1800d390d;
            }
            plVar11 = *(longlong **)(param_1 + 0x60);
            *(undefined8 *)(param_1 + 0x60) = 0;
            if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
              (**(code **)(plVar8[1] + 0x30))(plVar8);
            }
            uVar16 = 0x1ab;
          }
        }
      }
    }
  }
  else {
    iVar15 = *(int *)(lVar9 + 0xc);
    if (iVar15 == 0) {
      iVar15 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
      *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar15 + 1;
      *(int *)(lVar9 + 0xc) = iVar15;
    }
    if (iVar15 != DAT_1803a9770) {
      DAT_1803a9770 = iVar15;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#177="build_editorial_review_system_prompt" bytes=[3955753, 3955791] */
      DAT_1803b29d8 = FUN_1803707d0(lVar9,DAT_1803b3088,*(undefined8 *)(DAT_1803b3088 + 0x18));
    }
    if (-1 < DAT_1803b29d8) {
      lVar17 = (1L << (*(byte *)(lVar9 + 9) & 0x3f)) + lVar9;
      if (*(longlong *)(lVar17 + 0x28 + DAT_1803b29d8 * 0x10) != 0) goto LAB_1800d3478;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#177="build_editorial_review_system_prompt" bytes=[3955753, 3955791] */
      DAT_1803b29d8 = FUN_1803707d0(lVar9,DAT_1803b3088,*(undefined8 *)(DAT_1803b3088 + 0x18));
      if (-1 < DAT_1803b29d8) {
        lVar9 = *(longlong *)(lVar17 + 0x28 + DAT_1803b29d8 * 0x10);
        goto LAB_1800d3433;
      }
    }
LAB_1800d3438:
/* Analyst constant binding: novels.pipeline.imitation_short_runner#177="build_editorial_review_system_prompt" bytes=[3955753, 3955791] */
    plVar11 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803b3088);
    if ((plVar11 != (longlong *)0x0) && (*plVar11 != 0)) goto LAB_1800d3478;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#177="build_editorial_review_system_prompt" bytes=[3955753, 3955791] */
    FUN_180373340(param_1,&local_68,DAT_1803b3088);
    plVar11 = local_68;
LAB_1800d3499:
    uVar16 = 0x1a4;
    if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
      lVar9 = plVar10[1];
LAB_1800d369c:
      (**(code **)(lVar9 + 0x30))(plVar10);
    }
LAB_1800d36a2:
    if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
      (**(code **)(plVar8[1] + 0x30))(plVar8);
    }
    if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
      (**(code **)(plVar12[1] + 0x30))(plVar12);
    }
  }
LAB_1800d3929:
  piVar4 = (int *)plVar11[5];
  if (piVar4 == (int *)0x0) {
    pcVar13 = (code *)FUN_18038e2a0(plVar5,uVar16);
    plVar8 = (longlong *)plVar11[5];
    pcVar14 = pcVar13;
    if (pcVar13 == _Py_NoneStruct_exref) {
      pcVar14 = (code *)0x0;
    }
    else {
LAB_1800d3981:
      pcVar13 = pcVar14;
      if ((pcVar14 != (code *)0x0) && (*(int *)pcVar14 + 1 != 0)) {
        *(int *)pcVar14 = *(int *)pcVar14 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar4 + 6) == plVar5) goto LAB_1800d39ca;
    pcVar13 = (code *)FUN_18038e2a0(plVar5,uVar16);
    *(int **)(pcVar13 + 0x10) = piVar4;
    if (*piVar4 + 1 != 0) {
      *piVar4 = *piVar4 + 1;
    }
    plVar8 = (longlong *)plVar11[5];
    pcVar14 = pcVar13;
    if (pcVar13 != _Py_NoneStruct_exref) goto LAB_1800d3981;
    pcVar14 = (code *)0x0;
  }
  plVar11[5] = (longlong)pcVar14;
  if (((plVar8 != (longlong *)0x0) && (-1 < (int)*plVar8)) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)
     ) {
    (**(code **)(plVar8[1] + 0x30))();
  }
  if (((pcVar13 != (code *)0x0) && (-1 < *(int *)pcVar13)) &&
     (*(longlong *)pcVar13 = *(longlong *)pcVar13 + -1, *(longlong *)pcVar13 == 0)) {
    (**(code **)(*(longlong *)(pcVar13 + 8) + 0x30))(pcVar13);
  }
LAB_1800d39ca:
  plVar12 = local_50;
  FUN_18036d0a0(plVar5,&DAT_180391c1c,local_50,plVar1,plVar2,local_res20);
  plVar8 = DAT_1803b2a18;
  if (plVar5 == DAT_1803b2a18) {
    if ((-1 < (int)*DAT_1803b2a18) && (*DAT_1803b2a18 = *DAT_1803b2a18 + -1, *plVar8 == 0)) {
      (**(code **)(plVar8[1] + 0x30))();
    }
    DAT_1803b2a18 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if (((local_res20 != (longlong *)0x0) && (-1 < (int)*local_res20)) &&
     (*local_res20 = *local_res20 + -1, *local_res20 == 0)) {
    (**(code **)(local_res20[1] + 0x30))(local_res20);
  }
  if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
    (**(code **)(plVar12[1] + 0x30))(plVar12);
  }
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))(plVar1);
  }
  if ((-1 < (int)*plVar2) && (*plVar2 = *plVar2 + -1, *plVar2 == 0)) {
    (**(code **)(plVar2[1] + 0x30))(plVar2);
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong **)(param_1 + 0x60) = plVar11;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return (longlong *)0x0;
}