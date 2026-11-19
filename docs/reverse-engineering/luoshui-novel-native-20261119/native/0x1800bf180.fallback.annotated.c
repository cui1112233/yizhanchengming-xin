/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x1800bf180 */

/* WARNING: Function: __security_check_cookie replaced with injection: security_check_cookie */

longlong FUN_1800bf180(longlong param_1,longlong param_2,undefined8 *param_3)

{
  int *piVar1;
  undefined8 *puVar2;
  longlong *plVar3;
  longlong *plVar4;
  code *pcVar5;
  int iVar6;
  longlong lVar7;
  longlong *plVar8;
  undefined8 uVar9;
  longlong *plVar10;
  code *pcVar11;
  code *pcVar12;
  longlong lVar13;
  longlong *plVar14;
  undefined8 uVar15;
  longlong lVar16;
  longlong *plVar17;
  bool bVar18;
  undefined1 auStack_1b8 [32];
  code *local_198;
  code *local_190;
  code *local_188;
  longlong *local_180;
  longlong *local_178;
  longlong *local_170;
  longlong *local_168;
  longlong *local_160;
  undefined8 local_158;
  undefined8 local_150;
  undefined8 local_148;
  undefined8 local_140;
  undefined8 local_138;
  undefined8 local_130;
  longlong *local_128;
  longlong *local_120;
  undefined8 local_118;
  longlong local_110;
  longlong *local_108;
  longlong local_100;
  code *local_f8;
  longlong *local_f0;
  code *local_e8;
  longlong *local_e0;
  longlong *local_d8;
  longlong *local_d0;
  code *local_c8;
  code *local_c0;
  code *local_b8;
  longlong *local_b0;
  code *local_a8;
  longlong *local_a0;
  longlong *local_98;
  longlong local_90;
  longlong *local_88;
  code *local_80;
  code *local_78;
  code *local_70;
  longlong *local_68;
  longlong *local_60;
  longlong *local_58;
  longlong *local_50;
  longlong *local_48;
  ulonglong local_40;
  
  plVar3 = DAT_1803b1ef0;
  local_40 = DAT_1803a6040 ^ (ulonglong)auStack_1b8;
  plVar10 = (longlong *)*param_3;
  plVar17 = (longlong *)0x0;
  local_b0 = (longlong *)0x0;
  local_c8 = (code *)0x0;
  local_c0 = (code *)0x0;
  local_b8 = (code *)0x0;
  local_f0 = (longlong *)0x0;
  local_128 = (longlong *)0x0;
  local_e0 = (longlong *)0x0;
  local_d8 = (longlong *)0x0;
  local_d0 = (longlong *)0x0;
  local_110 = DAT_180391cf0;
  local_120 = plVar10;
  local_100 = param_2;
  if (DAT_1803b1ef0 == (longlong *)0x0) {
LAB_1800bf224:
    DAT_1803b1ef0 = (longlong *)FUN_18036cd10(param_1,DAT_1803b1ee8,DAT_1803ca7c0,0x80);
  }
  else if ((1 < *DAT_1803b1ef0) || (DAT_1803b1ef0[2] != 0)) {
    if ((DAT_1803b1ef0 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b1ef0 && (*DAT_1803b1ef0 = *DAT_1803b1ef0 + -1, *plVar3 == 0)))) {
      (**(code **)(plVar3[1] + 0x30))(plVar3);
    }
    goto LAB_1800bf224;
  }
  plVar3 = DAT_1803b1ef0;
  FUN_180001d00(param_1,DAT_1803b1ef0);
  lVar7 = *(longlong *)(*(longlong *)(param_2 + 0xe8) + 0x10);
  if (lVar7 == 0) {
/* novels.pipeline.chapter_review_cycle#101="audit_fn" */
    lVar7 = DAT_1803b2288 + 0x28;
/* novels.pipeline.chapter_review_cycle#101="audit_fn" */
    if (((byte)*(undefined4 *)(DAT_1803b2288 + 0x20) & 0x60) != 0x60) {
/* novels.pipeline.chapter_review_cycle#101="audit_fn" */
      lVar7 = *(longlong *)(DAT_1803b2288 + 0x30);
    }
    plVar17 = (longlong *)
              PyUnicode_FromFormat
                        ("cannot access free variable \'%s\' where it is not associated with a value in enclosing scope"
                         ,lVar7);
    lVar7 = FUN_180376770(param_1,*(undefined8 *)PyExc_NameError_exref,plVar17);
    if ((-1 < (int)*plVar17) && (*plVar17 = *plVar17 + -1, *plVar17 == 0)) {
      (**(code **)(plVar17[1] + 0x30))(plVar17);
    }
    FUN_18000d1d0(param_1,lVar7);
    uVar15 = 0x99;
    plVar17 = (longlong *)0x0;
  }
  else {
    uVar15 = 0x99;
    *(undefined4 *)(plVar3 + 5) = 0x99;
    local_a0 = (longlong *)FUN_180376770(param_1,lVar7,plVar10);
    if (local_a0 == (longlong *)0x0) {
      lVar7 = *(longlong *)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
    }
    else {
      local_b0 = local_a0;
/* novels.pipeline.chapter_review_cycle#61="callable" */
      plVar8 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803b2148);
      if ((plVar8 == (longlong *)0x0) || (*plVar8 == 0)) {
        PyErr_PrintEx(0);
        Py_Exit(1);
        pcVar11 = (code *)swi(3);
        lVar7 = (*pcVar11)();
        return lVar7;
      }
      if (*(longlong *)(*(longlong *)(param_2 + 0xf8) + 0x10) == 0) {
/* novels.pipeline.chapter_review_cycle#102="post_write_fn" */
        lVar7 = DAT_1803b2290 + 0x28;
/* novels.pipeline.chapter_review_cycle#102="post_write_fn" */
        if (((byte)*(undefined4 *)(DAT_1803b2290 + 0x20) & 0x60) != 0x60) {
/* novels.pipeline.chapter_review_cycle#102="post_write_fn" */
          lVar7 = *(longlong *)(DAT_1803b2290 + 0x30);
        }
        plVar8 = (longlong *)
                 PyUnicode_FromFormat
                           ("cannot access free variable \'%s\' where it is not associated with a value in enclosing scope"
                            ,lVar7);
        lVar7 = FUN_180376770(param_1,*(undefined8 *)PyExc_NameError_exref,plVar8);
        if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
          (**(code **)(plVar8[1] + 0x30))(plVar8);
        }
        FUN_18000d1d0(param_1,lVar7);
        uVar15 = 0x9a;
      }
      else {
        uVar15 = 0x9a;
        *(undefined4 *)(plVar3 + 5) = 0x9a;
        plVar8 = (longlong *)FUN_180376770(param_1);
        if (plVar8 == (longlong *)0x0) {
          lVar7 = *(longlong *)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
        }
        else {
          iVar6 = FUN_180001510(plVar8);
          if (iVar6 == -1) {
LAB_1800bf3e4:
            lVar7 = *(longlong *)(param_1 + 0x60);
            *(undefined8 *)(param_1 + 0x60) = 0;
            if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
              (**(code **)(plVar8[1] + 0x30))(plVar8);
            }
          }
          else {
            if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
              (**(code **)(plVar8[1] + 0x30))(plVar8);
            }
            if (iVar6 == 0) {
              local_a8 = _Py_NoneStruct_exref;
              if (*(int *)_Py_NoneStruct_exref + 1 != 0) {
                *(int *)_Py_NoneStruct_exref = *(int *)_Py_NoneStruct_exref + 1;
              }
LAB_1800bf4a1:
              local_c8 = local_a8;
/* novels.pipeline.chapter_review_cycle#61="callable" */
              uVar9 = FUN_18000d570(DAT_1803b2148);
              if (*(longlong *)(*(longlong *)(local_100 + 0xe0) + 0x10) == 0) {
/* novels.pipeline.chapter_review_cycle#103="ai_tells_fn" */
                FUN_180371050(param_1,&local_110,DAT_1803b2298);
                lVar7 = local_110;
                FUN_18000d1d0(param_1,local_110);
                uVar15 = 0x9b;
              }
              else {
                uVar15 = 0x9b;
                *(undefined4 *)(plVar3 + 5) = 0x9b;
                plVar8 = (longlong *)FUN_180376770(param_1,uVar9);
                if (plVar8 == (longlong *)0x0) {
                  lVar7 = *(longlong *)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                }
                else {
                  iVar6 = FUN_180001510(plVar8);
                  if (iVar6 == -1) goto LAB_1800bf3e4;
                  if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                    (**(code **)(plVar8[1] + 0x30))(plVar8);
                  }
                  if (iVar6 != 0) {
                    lVar7 = *(longlong *)(*(longlong *)(local_100 + 0xe0) + 0x10);
/* novels.pipeline.chapter_review_cycle#103="ai_tells_fn" */
                    lVar13 = DAT_1803b2298;
                    if (lVar7 == 0) goto LAB_1800bf44b;
                    *(undefined4 *)(plVar3 + 5) = 0x9b;
                    local_e8 = (code *)FUN_180376770(param_1,lVar7,plVar10);
                    if (local_e8 != (code *)0x0) goto LAB_1800bf5ac;
                    goto LAB_1800bf581;
                  }
                  local_e8 = _Py_NoneStruct_exref;
                  if (*(int *)_Py_NoneStruct_exref + 1 != 0) {
                    *(int *)_Py_NoneStruct_exref = *(int *)_Py_NoneStruct_exref + 1;
                  }
LAB_1800bf5ac:
                  pcVar11 = local_e8;
                  local_c0 = local_e8;
/* novels.pipeline.chapter_review_cycle#61="callable" */
                  uVar9 = FUN_18000d570(DAT_1803b2148);
                  if (*(longlong *)(*(longlong *)(local_100 + 0x100) + 0x10) == 0) {
/* novels.pipeline.chapter_review_cycle#104="sensitive_fn" */
                    FUN_180371050(param_1,&local_110,DAT_1803b22a0);
                    lVar7 = local_110;
                    FUN_18000d1d0(param_1,local_110);
                    uVar15 = 0x9c;
                    plVar17 = local_128;
                    goto LAB_1800bf8af;
                  }
                  uVar15 = 0x9c;
                  *(undefined4 *)(plVar3 + 5) = 0x9c;
                  plVar8 = (longlong *)FUN_180376770(param_1,uVar9);
                  if (plVar8 == (longlong *)0x0) goto LAB_1800bf581;
                  iVar6 = FUN_180001510(plVar8);
                  if (iVar6 == -1) {
                    lVar7 = *(longlong *)(param_1 + 0x60);
                    *(undefined8 *)(param_1 + 0x60) = 0;
                    plVar17 = local_128;
                    if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                      (**(code **)(plVar8[1] + 0x30))(plVar8);
                      plVar17 = local_128;
                    }
                    goto LAB_1800bf8af;
                  }
                  if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                    (**(code **)(plVar8[1] + 0x30))(plVar8);
                  }
                  if (iVar6 == 0) {
                    local_f8 = _Py_NoneStruct_exref;
                    if (*(int *)_Py_NoneStruct_exref + 1 != 0) {
                      *(int *)_Py_NoneStruct_exref = *(int *)_Py_NoneStruct_exref + 1;
                    }
LAB_1800bf6ff:
                    pcVar12 = local_f8;
                    lVar7 = *(longlong *)(DAT_1803ca7b8 + 0x20);
                    local_b8 = local_f8;
                    if (*(char *)(lVar7 + 10) == '\0') {
                      plVar10 = (longlong *)FUN_1800015a0();
                      if (plVar10 == (longlong *)0x0) {
                        lVar13 = 0;
                      }
                      else {
                        lVar13 = *plVar10;
                      }
LAB_1800bf7d3:
                      if (lVar13 == 0) goto LAB_1800bf7d8;
                    }
                    else {
                      iVar6 = *(int *)(lVar7 + 0xc);
                      if (iVar6 == 0) {
                        iVar6 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
                        *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar6 + 1;
                        *(int *)(lVar7 + 0xc) = iVar6;
                      }
                      if (iVar6 != DAT_1803a9740) {
                        DAT_1803a9740 = iVar6;
                        DAT_1803b1f38 =
/* novels.pipeline.chapter_review_cycle#105="_merge_deterministic_issues" */
                             FUN_1803707d0(lVar7,DAT_1803b22a8,*(undefined8 *)(DAT_1803b22a8 + 0x18)
                                          );
                      }
                      if (-1 < DAT_1803b1f38) {
                        lVar16 = (1L << (*(byte *)(lVar7 + 9) & 0x3f)) + lVar7;
                        lVar13 = *(longlong *)(lVar16 + 0x28 + DAT_1803b1f38 * 0x10);
                        if (lVar13 != 0) goto LAB_1800bf819;
                        DAT_1803b1f38 =
/* novels.pipeline.chapter_review_cycle#105="_merge_deterministic_issues" */
                             FUN_1803707d0(lVar7,DAT_1803b22a8,*(undefined8 *)(DAT_1803b22a8 + 0x18)
                                          );
                        if (-1 < DAT_1803b1f38) {
                          lVar13 = *(longlong *)(lVar16 + 0x28 + DAT_1803b1f38 * 0x10);
                          goto LAB_1800bf7d3;
                        }
                      }
LAB_1800bf7d8:
/* novels.pipeline.chapter_review_cycle#105="_merge_deterministic_issues" */
                      plVar10 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803b22a8);
                      if ((plVar10 == (longlong *)0x0) || (lVar13 = *plVar10, lVar13 == 0)) {
/* novels.pipeline.chapter_review_cycle#105="_merge_deterministic_issues" */
                        FUN_180373340(param_1,&local_110,DAT_1803b22a8);
                        uVar15 = 0x9d;
                        lVar7 = local_110;
                        plVar10 = local_120;
                        plVar17 = local_128;
                        goto LAB_1800bf8af;
                      }
                    }
LAB_1800bf819:
                    pcVar5 = local_a8;
                    uVar15 = 0x9d;
                    *(undefined4 *)(plVar3 + 5) = 0x9d;
                    local_88 = local_a0;
                    local_80 = local_a8;
                    local_78 = pcVar11;
                    local_70 = pcVar12;
                    local_108 = (longlong *)FUN_180379a80(param_1,lVar13,&local_88);
                    if (local_108 != (longlong *)0x0) {
                      local_f0 = local_108;
                      iVar6 = FUN_180001510(pcVar5);
                      if (iVar6 == -1) {
                        lVar7 = *(longlong *)(param_1 + 0x60);
                        uVar15 = 0x9e;
                        *(undefined8 *)(param_1 + 0x60) = 0;
                        plVar10 = local_120;
                        plVar17 = local_128;
                        goto LAB_1800bf8af;
                      }
                      if (iVar6 == 0) goto LAB_1800bfbf4;
                      plVar10 = (longlong *)FUN_18037f0c0();
                      if (plVar10 == (longlong *)0x0) goto LAB_1800bf896;
                      iVar6 = FUN_180001510(plVar10);
                      if (iVar6 == -1) {
                        lVar7 = *(longlong *)(param_1 + 0x60);
                        *(undefined8 *)(param_1 + 0x60) = 0;
                        if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                          (**(code **)(plVar10[1] + 0x30))(plVar10);
                        }
                      }
                      else {
                        if (iVar6 != 1) {
                          if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                            (**(code **)(plVar10[1] + 0x30))(plVar10);
                          }
                          plVar10 = (longlong *)FUN_180372160(param_1,0);
                        }
                        plVar17 = (longlong *)FUN_18000d280();
                        if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                          (**(code **)(plVar10[1] + 0x30))(plVar10);
                        }
                        if (plVar17 != (longlong *)0x0) {
                          local_118 = FUN_18036d750(plVar17);
                          local_190 = (code *)&local_118;
                          local_180 = (longlong *)0x48;
                          local_188 = (code *)0x1;
                          local_198 = (code *)DAT_1803b1e78;
                          plVar10 = (longlong *)
/* novels.pipeline.chapter_review_cycle#117="<genexpr>" */
                                    FUN_180367890(FUN_1800c0970,DAT_1803ca7c0,DAT_1803b2308,
/* novels.pipeline.chapter_review_cycle#118="run_chapter_review_cycle.<locals>.assess.<locals>.<genexpr>" */
                                                  DAT_1803b2310);
                          if ((-1 < (int)*plVar17) && (*plVar17 = *plVar17 + -1, *plVar17 == 0)) {
                            (**(code **)(plVar17[1] + 0x30))(plVar17);
                          }
                          plVar17 = (longlong *)FUN_180375a90(param_1,plVar10);
                          if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                            (**(code **)(plVar10[1] + 0x30))(plVar10);
                          }
                          if (plVar17 != (longlong *)0x0) {
                            local_128 = plVar17;
                            iVar6 = FUN_1803858f0(plVar17,DAT_1803ab250);
/* novels.pipeline.chapter_review_cycle#106="后写校验：" */
                            piVar1 = DAT_1803b22b0;
                            if (iVar6 == -1) {
                              lVar7 = *(longlong *)(param_1 + 0x60);
                              uVar15 = 0xa0;
                              *(undefined8 *)(param_1 + 0x60) = 0;
                              plVar10 = local_120;
                            }
                            else {
                              pcVar11 = local_e8;
                              if (iVar6 == 1) {
                                lVar7 = *(longlong *)(*(longlong *)(local_100 + 0xf0) + 0x10);
                                if (lVar7 == 0) {
/* novels.pipeline.chapter_review_cycle#41="log" */
                                  FUN_180371050(param_1,&local_110,DAT_1803b20a8);
                                  lVar7 = local_110;
                                  FUN_18000d1d0(param_1,local_110);
                                  uVar15 = 0xa1;
                                  plVar10 = local_120;
                                }
                                else {
                                  lVar13 = *(longlong *)(param_1 + 0x10);
                                  plVar8 = *(longlong **)(lVar13 + 0x41da8);
                                  if (plVar8 == (longlong *)0x0) {
                                    plVar8 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,3);
                                  }
                                  else {
                                    *(longlong *)(lVar13 + 0x41da8) = plVar8[3];
                                    *(int *)(lVar13 + 0x41e40) = *(int *)(lVar13 + 0x41e40) + -1;
                                    *plVar8 = 1;
                                  }
                                  plVar10 = plVar8 + -2;
                                  plVar8[3] = 0;
                                  plVar8[4] = 0;
                                  plVar8[5] = 0;
                                  lVar13 = _PyThreadState_GetCurrent();
                                  lVar13 = *(longlong *)(*(longlong *)(lVar13 + 0x10) + 0xd0);
                                  puVar2 = *(undefined8 **)(lVar13 + 8);
                                  *puVar2 = plVar10;
                                  plVar8[-1] = plVar8[-1] & 3;
                                  plVar8[-1] = plVar8[-1] | (ulonglong)puVar2;
                                  *plVar10 = lVar13;
                                  *(longlong **)(lVar13 + 8) = plVar10;
                                  if (*piVar1 + 1 != 0) {
                                    *piVar1 = *piVar1 + 1;
                                  }
                                  plVar8[3] = (longlong)piVar1;
/* novels.pipeline.chapter_review_cycle#18="" */
                                  lVar13 = FUN_180375520(param_1,plVar17,DAT_1803b1ff0);
                                  if (lVar13 != 0) {
                                    plVar8[4] = lVar13;
/* novels.pipeline.chapter_review_cycle#107=" 个硬错误" */
                                    piVar1 = DAT_1803b22b8;
/* novels.pipeline.chapter_review_cycle#107=" 个硬错误" */
                                    if (*DAT_1803b22b8 + 1 != 0) {
/* novels.pipeline.chapter_review_cycle#107=" 个硬错误" */
                                      *DAT_1803b22b8 = *DAT_1803b22b8 + 1;
                                    }
                                    plVar8[5] = (longlong)piVar1;
/* novels.pipeline.chapter_review_cycle#18="" */
                                    plVar10 = (longlong *)PyUnicode_Join(DAT_1803b1ff0,plVar8);
                                    if ((-1 < (int)*plVar8) &&
                                       (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                      (**(code **)(plVar8[1] + 0x30))(plVar8);
                                    }
                                    uVar15 = 0xa1;
                                    if (plVar10 != (longlong *)0x0) {
                                      *(undefined4 *)(plVar3 + 5) = 0xa1;
                                      plVar17 = (longlong *)FUN_180376770(param_1,lVar7,plVar10);
                                      if ((-1 < (int)*plVar10) &&
                                         (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                        (**(code **)(plVar10[1] + 0x30))(plVar10);
                                      }
                                      if (plVar17 != (longlong *)0x0) {
                                        pcVar12 = local_f8;
                                        pcVar11 = local_e8;
                                        if ((-1 < (int)*plVar17) &&
                                           (*plVar17 = *plVar17 + -1, *plVar17 == 0)) {
                                          (**(code **)(plVar17[1] + 0x30))(plVar17);
                                          pcVar12 = local_f8;
                                          pcVar11 = local_e8;
                                        }
                                        goto LAB_1800bfbf4;
                                      }
                                    }
                                    goto LAB_1800bf6d6;
                                  }
                                  lVar7 = *(longlong *)(param_1 + 0x60);
                                  uVar15 = 0xa1;
                                  *(undefined8 *)(param_1 + 0x60) = 0;
                                  plVar10 = local_120;
                                  if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0))
                                  {
                                    (**(code **)(plVar8[1] + 0x30))(plVar8);
                                    plVar10 = local_120;
                                  }
                                }
                              }
                              else {
LAB_1800bfbf4:
                                iVar6 = FUN_180001510(pcVar11);
                                if (iVar6 == -1) {
                                  lVar7 = *(longlong *)(param_1 + 0x60);
                                  uVar15 = 0xa2;
                                  *(undefined8 *)(param_1 + 0x60) = 0;
                                  plVar10 = local_120;
                                  plVar17 = local_128;
                                }
                                else {
                                  bVar18 = iVar6 != 0;
                                  if (iVar6 == 0) {
LAB_1800bfcaa:
/* novels.pipeline.chapter_review_cycle#108="AI味检测：" */
                                    piVar1 = DAT_1803b22c0;
                                    if (bVar18) {
                                      lVar7 = *(longlong *)(*(longlong *)(local_100 + 0xf0) + 0x10);
                                      if (lVar7 == 0) {
/* novels.pipeline.chapter_review_cycle#41="log" */
                                        FUN_180371050(param_1,&local_110,DAT_1803b20a8);
                                        lVar7 = local_110;
                                        FUN_18000d1d0(param_1,local_110);
                                        uVar15 = 0xa3;
                                        plVar10 = local_120;
                                        plVar17 = local_128;
                                      }
                                      else {
                                        lVar13 = *(longlong *)(param_1 + 0x10);
                                        plVar8 = *(longlong **)(lVar13 + 0x41da8);
                                        if (plVar8 == (longlong *)0x0) {
                                          plVar8 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,3);
                                        }
                                        else {
                                          *(longlong *)(lVar13 + 0x41da8) = plVar8[3];
                                          *(int *)(lVar13 + 0x41e40) =
                                               *(int *)(lVar13 + 0x41e40) + -1;
                                          *plVar8 = 1;
                                        }
                                        plVar10 = plVar8 + -2;
                                        plVar8[3] = 0;
                                        plVar8[4] = 0;
                                        plVar8[5] = 0;
                                        lVar13 = _PyThreadState_GetCurrent();
                                        lVar13 = *(longlong *)(*(longlong *)(lVar13 + 0x10) + 0xd0);
                                        puVar2 = *(undefined8 **)(lVar13 + 8);
                                        *puVar2 = plVar10;
                                        plVar8[-1] = plVar8[-1] & 3;
                                        plVar8[-1] = plVar8[-1] | (ulonglong)puVar2;
                                        *plVar10 = lVar13;
                                        *(longlong **)(lVar13 + 8) = plVar10;
                                        if (*piVar1 + 1 != 0) {
                                          *piVar1 = *piVar1 + 1;
                                        }
                                        plVar8[3] = (longlong)piVar1;
                                        plVar10 = (longlong *)
/* novels.pipeline.chapter_review_cycle#24="issues" */
                                                  FUN_18037f0c0(puVar2,local_e8,DAT_1803b2020);
                                        if (plVar10 != (longlong *)0x0) {
                                          lVar13 = FUN_180373860(plVar10);
                                          if ((lVar13 < 0) && (*(longlong *)(param_1 + 0x60) != 0))
                                          {
                                            plVar17 = (longlong *)0x0;
                                          }
                                          else {
                                            plVar17 = (longlong *)PyLong_FromSsize_t(lVar13);
                                          }
                                          if ((-1 < (int)*plVar10) &&
                                             (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                            (**(code **)(plVar10[1] + 0x30))(plVar10);
                                          }
                                          if (plVar17 != (longlong *)0x0) {
/* novels.pipeline.chapter_review_cycle#18="" */
                                            lVar13 = FUN_180375520(param_1,plVar17,DAT_1803b1ff0);
                                            if ((-1 < (int)*plVar17) &&
                                               (*plVar17 = *plVar17 + -1, *plVar17 == 0)) {
                                              (**(code **)(plVar17[1] + 0x30))(plVar17);
                                            }
                                            if (lVar13 != 0) {
                                              plVar8[4] = lVar13;
/* novels.pipeline.chapter_review_cycle#109=" 个问题" */
                                              piVar1 = DAT_1803b22c8;
/* novels.pipeline.chapter_review_cycle#109=" 个问题" */
                                              if (*DAT_1803b22c8 + 1 != 0) {
/* novels.pipeline.chapter_review_cycle#109=" 个问题" */
                                                *DAT_1803b22c8 = *DAT_1803b22c8 + 1;
                                              }
                                              plVar8[5] = (longlong)piVar1;
                                              plVar10 = (longlong *)
/* novels.pipeline.chapter_review_cycle#18="" */
                                                        PyUnicode_Join(DAT_1803b1ff0,plVar8);
                                              if ((-1 < (int)*plVar8) &&
                                                 (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                (**(code **)(plVar8[1] + 0x30))(plVar8);
                                              }
                                              uVar15 = 0xa3;
                                              if (plVar10 != (longlong *)0x0) {
                                                *(undefined4 *)(plVar3 + 5) = 0xa3;
                                                plVar17 = (longlong *)
                                                          FUN_180376770(param_1,lVar7,plVar10);
                                                if ((-1 < (int)*plVar10) &&
                                                   (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                                  (**(code **)(plVar10[1] + 0x30))(plVar10);
                                                }
                                                if (plVar17 != (longlong *)0x0) {
                                                  if ((-1 < (int)*plVar17) &&
                                                     (*plVar17 = *plVar17 + -1, *plVar17 == 0)) {
                                                    (**(code **)(plVar17[1] + 0x30))(plVar17);
                                                  }
                                                  goto LAB_1800bfee5;
                                                }
                                              }
                                              goto LAB_1800bf6d6;
                                            }
                                          }
                                        }
                                        lVar7 = *(longlong *)(param_1 + 0x60);
                                        uVar15 = 0xa3;
                                        *(undefined8 *)(param_1 + 0x60) = 0;
LAB_1800bfe1b:
                                        plVar10 = local_120;
                                        plVar17 = local_128;
                                        if ((-1 < (int)*plVar8) &&
                                           (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                          (**(code **)(plVar8[1] + 0x30))(plVar8);
                                          plVar10 = local_120;
                                          plVar17 = local_128;
                                        }
                                      }
                                    }
                                    else {
LAB_1800bfee5:
                                      iVar6 = FUN_180001510(pcVar12);
                                      if (iVar6 == -1) {
                                        lVar7 = *(longlong *)(param_1 + 0x60);
                                        uVar15 = 0xa4;
                                        *(undefined8 *)(param_1 + 0x60) = 0;
                                        plVar10 = local_120;
                                        plVar17 = local_128;
                                      }
                                      else {
                                        bVar18 = iVar6 != 0;
                                        if (iVar6 == 0) {
LAB_1800bff94:
/* novels.pipeline.chapter_review_cycle#110="敏感词检测：" */
                                          piVar1 = DAT_1803b22d0;
                                          if (bVar18) {
                                            lVar7 = *(longlong *)
                                                     (*(longlong *)(local_100 + 0xf0) + 0x10);
                                            if (lVar7 != 0) {
                                              lVar13 = *(longlong *)(param_1 + 0x10);
                                              plVar8 = *(longlong **)(lVar13 + 0x41da8);
                                              if (plVar8 == (longlong *)0x0) {
                                                plVar8 = (longlong *)
                                                         FUN_18035ebd0(PyTuple_Type_exref,3);
                                              }
                                              else {
                                                *(longlong *)(lVar13 + 0x41da8) = plVar8[3];
                                                *(int *)(lVar13 + 0x41e40) =
                                                     *(int *)(lVar13 + 0x41e40) + -1;
                                                *plVar8 = 1;
                                              }
                                              plVar10 = plVar8 + -2;
                                              plVar8[3] = 0;
                                              plVar8[4] = 0;
                                              plVar8[5] = 0;
                                              lVar13 = _PyThreadState_GetCurrent();
                                              lVar13 = *(longlong *)
                                                        (*(longlong *)(lVar13 + 0x10) + 0xd0);
                                              puVar2 = *(undefined8 **)(lVar13 + 8);
                                              *puVar2 = plVar10;
                                              plVar8[-1] = plVar8[-1] & 3;
                                              plVar8[-1] = plVar8[-1] | (ulonglong)puVar2;
                                              *plVar10 = lVar13;
                                              *(longlong **)(lVar13 + 8) = plVar10;
                                              if (*piVar1 + 1 != 0) {
                                                *piVar1 = *piVar1 + 1;
                                              }
                                              plVar8[3] = (longlong)piVar1;
                                              plVar10 = (longlong *)
/* novels.pipeline.chapter_review_cycle#24="issues" */
                                                        FUN_18037f0c0(puVar2,pcVar12,DAT_1803b2020);
                                              if (plVar10 == (longlong *)0x0) {
                                                lVar7 = *(longlong *)(param_1 + 0x60);
                                                *(undefined8 *)(param_1 + 0x60) = 0;
                                                uVar15 = 0xa5;
                                                goto LAB_1800bfe1b;
                                              }
                                              lVar13 = FUN_180373860(plVar10);
                                              if ((lVar13 < 0) &&
                                                 (*(longlong *)(param_1 + 0x60) != 0)) {
                                                plVar17 = (longlong *)0x0;
                                              }
                                              else {
                                                plVar17 = (longlong *)PyLong_FromSsize_t(lVar13);
                                              }
                                              if ((-1 < (int)*plVar10) &&
                                                 (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                                (**(code **)(plVar10[1] + 0x30))(plVar10);
                                              }
                                              if (plVar17 != (longlong *)0x0) {
/* novels.pipeline.chapter_review_cycle#18="" */
                                                lVar13 = FUN_180375520(param_1,plVar17,DAT_1803b1ff0
                                                                      );
                                                if ((-1 < (int)*plVar17) &&
                                                   (*plVar17 = *plVar17 + -1, *plVar17 == 0)) {
                                                  (**(code **)(plVar17[1] + 0x30))(plVar17);
                                                }
                                                if (lVar13 != 0) {
                                                  plVar8[4] = lVar13;
/* novels.pipeline.chapter_review_cycle#109=" 个问题" */
                                                  piVar1 = DAT_1803b22c8;
/* novels.pipeline.chapter_review_cycle#109=" 个问题" */
                                                  if (*DAT_1803b22c8 + 1 != 0) {
/* novels.pipeline.chapter_review_cycle#109=" 个问题" */
                                                    *DAT_1803b22c8 = *DAT_1803b22c8 + 1;
                                                  }
                                                  plVar8[5] = (longlong)piVar1;
                                                  plVar10 = (longlong *)
/* novels.pipeline.chapter_review_cycle#18="" */
                                                            PyUnicode_Join(DAT_1803b1ff0,plVar8);
                                                  if ((-1 < (int)*plVar8) &&
                                                     (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                    (**(code **)(plVar8[1] + 0x30))(plVar8);
                                                  }
                                                  uVar15 = 0xa5;
                                                  if (plVar10 == (longlong *)0x0) {
                                                    lVar7 = *(longlong *)(param_1 + 0x60);
                                                    *(undefined8 *)(param_1 + 0x60) = 0;
                                                    plVar10 = local_120;
                                                    plVar17 = local_128;
                                                  }
                                                  else {
                                                    *(undefined4 *)(plVar3 + 5) = 0xa5;
                                                    plVar17 = (longlong *)
                                                              FUN_180376770(param_1,lVar7,plVar10);
                                                    if ((-1 < (int)*plVar10) &&
                                                       (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                                      (**(code **)(plVar10[1] + 0x30))(plVar10);
                                                    }
                                                    if (plVar17 != (longlong *)0x0) {
                                                      if ((-1 < (int)*plVar17) &&
                                                         (*plVar17 = *plVar17 + -1, *plVar17 == 0))
                                                      {
                                                        (**(code **)(plVar17[1] + 0x30))(plVar17);
                                                      }
                                                      goto LAB_1800c01d7;
                                                    }
                                                    lVar7 = *(longlong *)(param_1 + 0x60);
                                                    *(undefined8 *)(param_1 + 0x60) = 0;
                                                    plVar10 = local_120;
                                                    plVar17 = local_128;
                                                  }
                                                  goto LAB_1800bf8af;
                                                }
                                              }
                                              lVar7 = *(longlong *)(param_1 + 0x60);
                                              uVar15 = 0xa5;
                                              *(undefined8 *)(param_1 + 0x60) = 0;
                                              goto LAB_1800bfe1b;
                                            }
/* novels.pipeline.chapter_review_cycle#41="log" */
                                            FUN_180371050(param_1,&local_110,DAT_1803b20a8);
                                            lVar7 = local_110;
                                            FUN_18000d1d0(param_1,local_110);
                                            uVar15 = 0xa5;
                                            plVar10 = local_120;
                                            plVar17 = local_128;
                                          }
                                          else {
LAB_1800c01d7:
                                            lVar7 = *(longlong *)(DAT_1803ca7b8 + 0x20);
                                            if (*(char *)(lVar7 + 10) == '\0') {
                                              plVar10 = (longlong *)
/* novels.pipeline.chapter_review_cycle#111="_count_chars" */
                                                        FUN_1800015a0(DAT_1803ca7b8,DAT_1803b22d8);
                                              lVar13 = 0;
                                              if (plVar10 != (longlong *)0x0) {
                                                lVar13 = *plVar10;
                                              }
LAB_1800c02a2:
                                              if (lVar13 == 0) goto LAB_1800c02a7;
                                            }
                                            else {
                                              iVar6 = *(int *)(lVar7 + 0xc);
                                              if (iVar6 == 0) {
                                                iVar6 = *(int *)(*(longlong *)(param_1 + 0x10) +
                                                                0x42118);
                                                *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) =
                                                     iVar6 + 1;
                                                *(int *)(lVar7 + 0xc) = iVar6;
                                              }
                                              if (iVar6 != DAT_1803a975c) {
                                                DAT_1803a975c = iVar6;
                                                DAT_1803b1f50 =
/* novels.pipeline.chapter_review_cycle#111="_count_chars" */
                                                     FUN_1803707d0(lVar7,DAT_1803b22d8,
                                                                   *(undefined8 *)
/* novels.pipeline.chapter_review_cycle#111="_count_chars" */
                                                                    (DAT_1803b22d8 + 0x18));
                                              }
                                              if (-1 < DAT_1803b1f50) {
                                                lVar16 = (1L << (*(byte *)(lVar7 + 9) & 0x3f)) +
                                                         lVar7;
                                                lVar13 = *(longlong *)
                                                          (lVar16 + 0x28 + DAT_1803b1f50 * 0x10);
                                                if (lVar13 != 0) goto LAB_1800c02e8;
                                                DAT_1803b1f50 =
/* novels.pipeline.chapter_review_cycle#111="_count_chars" */
                                                     FUN_1803707d0(lVar7,DAT_1803b22d8,
                                                                   *(undefined8 *)
/* novels.pipeline.chapter_review_cycle#111="_count_chars" */
                                                                    (DAT_1803b22d8 + 0x18));
                                                if (-1 < DAT_1803b1f50) {
                                                  lVar13 = *(longlong *)
                                                            (lVar16 + 0x28 + DAT_1803b1f50 * 0x10);
                                                  goto LAB_1800c02a2;
                                                }
                                              }
LAB_1800c02a7:
                                              plVar10 = (longlong *)
/* novels.pipeline.chapter_review_cycle#111="_count_chars" */
                                                        FUN_1800015a0(DAT_1803ab778,DAT_1803b22d8);
                                              if ((plVar10 == (longlong *)0x0) ||
                                                 (lVar13 = *plVar10, lVar13 == 0)) {
/* novels.pipeline.chapter_review_cycle#111="_count_chars" */
                                                FUN_180373340(param_1,&local_110,DAT_1803b22d8);
                                                uVar15 = 0xa6;
                                                lVar7 = local_110;
                                                plVar10 = local_120;
                                                plVar17 = local_128;
                                                goto LAB_1800bf8af;
                                              }
                                            }
LAB_1800c02e8:
                                            plVar10 = local_120;
                                            uVar15 = 0xa6;
                                            *(undefined4 *)(plVar3 + 5) = 0xa6;
                                            plVar17 = (longlong *)
                                                      FUN_180376770(param_1,lVar13,local_120);
                                            if (plVar17 == (longlong *)0x0) {
                                              lVar7 = *(longlong *)(param_1 + 0x60);
                                              *(undefined8 *)(param_1 + 0x60) = 0;
                                              plVar17 = local_128;
                                            }
                                            else {
                                              lVar7 = *(longlong *)(DAT_1803ca7b8 + 0x20);
                                              local_e0 = plVar17;
                                              if (*(char *)(lVar7 + 10) == '\0') {
                                                plVar8 = (longlong *)
/* novels.pipeline.chapter_review_cycle#112="_audit_score" */
                                                         FUN_1800015a0(DAT_1803ca7b8,DAT_1803b22e0);
                                                lVar13 = 0;
                                                if (plVar8 != (longlong *)0x0) {
                                                  lVar13 = *plVar8;
                                                }
LAB_1800c03e5:
                                                if (lVar13 == 0) goto LAB_1800c03ea;
                                              }
                                              else {
                                                iVar6 = *(int *)(lVar7 + 0xc);
                                                if (iVar6 == 0) {
                                                  iVar6 = *(int *)(*(longlong *)(param_1 + 0x10) +
                                                                  0x42118);
                                                  *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118)
                                                       = iVar6 + 1;
                                                  *(int *)(lVar7 + 0xc) = iVar6;
                                                }
                                                if (iVar6 != DAT_1803a973c) {
                                                  DAT_1803a973c = iVar6;
                                                  DAT_1803b1ec0 =
/* novels.pipeline.chapter_review_cycle#112="_audit_score" */
                                                       FUN_1803707d0(lVar7,DAT_1803b22e0,
                                                                     *(undefined8 *)
/* novels.pipeline.chapter_review_cycle#112="_audit_score" */
                                                                      (DAT_1803b22e0 + 0x18));
                                                }
                                                if (-1 < DAT_1803b1ec0) {
                                                  lVar16 = (1L << (*(byte *)(lVar7 + 9) & 0x3f)) +
                                                           lVar7;
                                                  lVar13 = *(longlong *)
                                                            (lVar16 + 0x28 + DAT_1803b1ec0 * 0x10);
                                                  if (lVar13 != 0) goto LAB_1800c042f;
                                                  DAT_1803b1ec0 =
/* novels.pipeline.chapter_review_cycle#112="_audit_score" */
                                                       FUN_1803707d0(lVar7,DAT_1803b22e0,
                                                                     *(undefined8 *)
/* novels.pipeline.chapter_review_cycle#112="_audit_score" */
                                                                      (DAT_1803b22e0 + 0x18));
                                                  if (-1 < DAT_1803b1ec0) {
                                                    lVar13 = *(longlong *)
                                                              (lVar16 + 0x28 + DAT_1803b1ec0 * 0x10)
                                                    ;
                                                    goto LAB_1800c03e5;
                                                  }
                                                }
LAB_1800c03ea:
                                                plVar8 = (longlong *)
/* novels.pipeline.chapter_review_cycle#112="_audit_score" */
                                                         FUN_1800015a0(DAT_1803ab778,DAT_1803b22e0);
                                                if ((plVar8 == (longlong *)0x0) ||
                                                   (lVar13 = *plVar8, lVar13 == 0)) {
/* novels.pipeline.chapter_review_cycle#112="_audit_score" */
                                                  FUN_180373340(param_1,&local_110,DAT_1803b22e0);
                                                  uVar15 = 0xa7;
                                                  lVar7 = local_110;
                                                  plVar17 = local_128;
                                                  goto LAB_1800bf8af;
                                                }
                                              }
LAB_1800c042f:
                                              uVar15 = 0xa7;
                                              *(undefined4 *)(plVar3 + 5) = 0xa7;
                                              plVar8 = (longlong *)
                                                       FUN_180376770(param_1,lVar13,local_f0);
                                              if (plVar8 == (longlong *)0x0) {
LAB_1800bf581:
                                                lVar7 = *(longlong *)(param_1 + 0x60);
                                                *(undefined8 *)(param_1 + 0x60) = 0;
                                                plVar17 = local_128;
                                              }
                                              else {
                                                local_d8 = plVar8;
                                                lVar7 = FUN_1800b50d0(param_1);
                                                if (lVar7 == 0) {
/* novels.pipeline.chapter_review_cycle#113="_length_in_range" */
                                                  FUN_180373340(param_1,&local_110,DAT_1803b22e8);
                                                  uVar15 = 0xa8;
                                                  lVar7 = local_110;
                                                  plVar17 = local_128;
                                                }
                                                else {
                                                  lVar13 = *(longlong *)
                                                            (*(longlong *)(local_100 + 0x108) + 0x10
                                                            );
                                                  if (lVar13 == 0) {
/* novels.pipeline.chapter_review_cycle#114="target_word_count" */
                                                    FUN_180371050(param_1,&local_110,DAT_1803b22f0);
                                                    lVar7 = local_110;
                                                    FUN_18000d1d0(param_1,local_110);
                                                    uVar15 = 0xa8;
                                                    plVar17 = local_128;
                                                  }
                                                  else {
                                                    uVar15 = 0xa8;
                                                    *(undefined4 *)(plVar3 + 5) = 0xa8;
                                                    local_98 = plVar17;
                                                    local_90 = lVar13;
                                                    plVar14 = (longlong *)
                                                              FUN_180377750(param_1,lVar7,&local_98)
                                                    ;
                                                    if (plVar14 == (longlong *)0x0)
                                                    goto LAB_1800bf581;
                                                    local_d0 = plVar14;
                                                    lVar7 = FUN_1800b4d00(param_1);
                                                    if (lVar7 != 0) {
                                                      local_58 = local_f0;
                                                      uVar15 = 0xa9;
                                                      *(undefined4 *)(plVar3 + 5) = 0xa9;
                                                      local_68 = plVar10;
                                                      local_60 = plVar17;
                                                      local_50 = plVar8;
                                                      local_48 = plVar14;
                                                      lVar7 = FUN_18037ccb0(param_1,lVar7,&local_68,
                                                                            DAT_1803b2300);
                                                      if (lVar7 != 0) {
                                                        FUN_180001d80(param_1);
                                                        if ((-1 < (int)*local_a0) &&
                                                           (*local_a0 = *local_a0 + -1,
                                                           *local_a0 == 0)) {
                                                          (**(code **)(local_b0[1] + 0x30))();
                                                        }
                                                        if ((-1 < *(int *)local_a8) &&
                                                           (*(longlong *)local_a8 =
                                                                 *(longlong *)local_a8 + -1,
                                                           *(longlong *)local_a8 == 0)) {
                                                          (**(code **)(*(longlong *)(local_a8 + 8) +
                                                                      0x30))();
                                                        }
                                                        if ((-1 < *(int *)local_e8) &&
                                                           (*(longlong *)local_e8 =
                                                                 *(longlong *)local_e8 + -1,
                                                           *(longlong *)local_e8 == 0)) {
                                                          (**(code **)(*(longlong *)(local_e8 + 8) +
                                                                      0x30))();
                                                        }
                                                        if ((-1 < *(int *)local_f8) &&
                                                           (*(longlong *)local_f8 =
                                                                 *(longlong *)local_f8 + -1,
                                                           *(longlong *)local_f8 == 0)) {
                                                          (**(code **)(*(longlong *)(local_f8 + 8) +
                                                                      0x30))();
                                                        }
                                                        if ((-1 < (int)*local_108) &&
                                                           (*local_108 = *local_108 + -1,
                                                           *local_108 == 0)) {
                                                          (**(code **)(local_f0[1] + 0x30))();
                                                        }
                                                        if (((local_128 != (longlong *)0x0) &&
                                                            (-1 < (int)*local_128)) &&
                                                           (*local_128 = *local_128 + -1,
                                                           *local_128 == 0)) {
                                                          (**(code **)(local_128[1] + 0x30))();
                                                        }
                                                        if ((-1 < (int)*plVar17) &&
                                                           (*plVar17 = *plVar17 + -1, *plVar17 == 0)
                                                           ) {
                                                          (**(code **)(plVar17[1] + 0x30))(plVar17);
                                                        }
                                                        if ((-1 < (int)*plVar8) &&
                                                           (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                          (**(code **)(plVar8[1] + 0x30))(plVar8);
                                                        }
                                                        if ((-1 < (int)*plVar14) &&
                                                           (*plVar14 = *plVar14 + -1, *plVar14 == 0)
                                                           ) {
                                                          (**(code **)(plVar14[1] + 0x30))(plVar14);
                                                        }
                                                        if ((int)*local_120 < 0) {
                                                          return lVar7;
                                                        }
                                                        *local_120 = *local_120 + -1;
                                                        if (*local_120 != 0) {
                                                          return lVar7;
                                                        }
                                                        (**(code **)(local_120[1] + 0x30))();
                                                        return lVar7;
                                                      }
                                                      goto LAB_1800bf6d6;
                                                    }
/* novels.pipeline.chapter_review_cycle#115="ReviewSnapshot" */
                                                    FUN_180373340(param_1,&local_110,DAT_1803b22f8);
                                                    uVar15 = 0xa9;
                                                    lVar7 = local_110;
                                                    plVar17 = local_128;
                                                  }
                                                }
                                              }
                                            }
                                          }
                                        }
                                        else {
                                          plVar10 = (longlong *)FUN_18037f0c0();
                                          if (plVar10 == (longlong *)0x0) {
                                            lVar7 = *(longlong *)(param_1 + 0x60);
                                            uVar15 = 0xa4;
                                            *(undefined8 *)(param_1 + 0x60) = 0;
                                            plVar10 = local_120;
                                            plVar17 = local_128;
                                          }
                                          else {
                                            iVar6 = FUN_180001510(plVar10);
                                            if (iVar6 != -1) {
                                              bVar18 = iVar6 != 0;
                                              if ((-1 < (int)*plVar10) &&
                                                 (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                                (**(code **)(plVar10[1] + 0x30))(plVar10);
                                              }
                                              goto LAB_1800bff94;
                                            }
                                            lVar7 = *(longlong *)(param_1 + 0x60);
                                            *(undefined8 *)(param_1 + 0x60) = 0;
                                            if ((-1 < (int)*plVar10) &&
                                               (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                              (**(code **)(plVar10[1] + 0x30))(plVar10);
                                            }
                                            uVar15 = 0xa4;
                                            plVar10 = local_120;
                                            plVar17 = local_128;
                                          }
                                        }
                                      }
                                    }
                                  }
                                  else {
                                    plVar10 = (longlong *)FUN_18037f0c0();
                                    if (plVar10 == (longlong *)0x0) {
                                      lVar7 = *(longlong *)(param_1 + 0x60);
                                      uVar15 = 0xa2;
                                      *(undefined8 *)(param_1 + 0x60) = 0;
                                      plVar10 = local_120;
                                      plVar17 = local_128;
                                    }
                                    else {
                                      iVar6 = FUN_180001510(plVar10);
                                      if (iVar6 != -1) {
                                        bVar18 = iVar6 != 0;
                                        if ((-1 < (int)*plVar10) &&
                                           (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                          (**(code **)(plVar10[1] + 0x30))(plVar10);
                                        }
                                        goto LAB_1800bfcaa;
                                      }
                                      lVar7 = *(longlong *)(param_1 + 0x60);
                                      *(undefined8 *)(param_1 + 0x60) = 0;
                                      if ((-1 < (int)*plVar10) &&
                                         (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                        (**(code **)(plVar10[1] + 0x30))(plVar10);
                                      }
                                      uVar15 = 0xa2;
                                      plVar10 = local_120;
                                      plVar17 = local_128;
                                    }
                                  }
                                }
                              }
                            }
                            goto LAB_1800bf8af;
                          }
                        }
LAB_1800bf896:
                        lVar7 = *(longlong *)(param_1 + 0x60);
                        *(undefined8 *)(param_1 + 0x60) = 0;
                      }
                      uVar15 = 0x9f;
                      plVar10 = local_120;
                      plVar17 = local_128;
                      goto LAB_1800bf8af;
                    }
                  }
                  else {
                    lVar7 = *(longlong *)(*(longlong *)(local_100 + 0x100) + 0x10);
                    if (lVar7 == 0) {
/* novels.pipeline.chapter_review_cycle#104="sensitive_fn" */
                      FUN_180371050(param_1,&local_110,DAT_1803b22a0);
                      lVar7 = local_110;
                      FUN_18000d1d0(param_1,local_110);
                      plVar17 = local_128;
                      goto LAB_1800bf8af;
                    }
                    *(undefined4 *)(plVar3 + 5) = 0x9c;
                    local_f8 = (code *)FUN_180376770(param_1,lVar7,plVar10);
                    if (local_f8 != (code *)0x0) goto LAB_1800bf6ff;
                  }
LAB_1800bf6d6:
                  lVar7 = *(longlong *)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                  plVar10 = local_120;
                  plVar17 = local_128;
                }
              }
            }
            else {
              lVar7 = *(longlong *)(*(longlong *)(local_100 + 0xf8) + 0x10);
/* novels.pipeline.chapter_review_cycle#102="post_write_fn" */
              lVar13 = DAT_1803b2290;
              if (lVar7 != 0) {
                *(undefined4 *)(plVar3 + 5) = 0x9a;
                local_a8 = (code *)FUN_180376770(param_1,lVar7,plVar10);
                if (local_a8 == (code *)0x0) {
                  lVar7 = *(longlong *)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                  goto LAB_1800bf8af;
                }
                goto LAB_1800bf4a1;
              }
LAB_1800bf44b:
              FUN_180371050(param_1,&local_110,lVar13);
              lVar7 = local_110;
              FUN_18000d1d0(param_1,local_110);
            }
          }
        }
      }
    }
  }
LAB_1800bf8af:
  piVar1 = *(int **)(lVar7 + 0x28);
  if (piVar1 == (int *)0x0) {
    pcVar11 = (code *)FUN_18038e2a0(plVar3,uVar15);
    plVar8 = *(longlong **)(lVar7 + 0x28);
    pcVar12 = pcVar11;
    if (pcVar11 == _Py_NoneStruct_exref) {
      pcVar12 = (code *)0x0;
    }
    else {
LAB_1800c06da:
      pcVar11 = pcVar12;
      if ((pcVar12 != (code *)0x0) && (*(int *)pcVar12 + 1 != 0)) {
        *(int *)pcVar12 = *(int *)pcVar12 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar1 + 6) == plVar3) goto LAB_1800c0723;
    pcVar11 = (code *)FUN_18038e2a0(plVar3,uVar15);
    *(int **)(pcVar11 + 0x10) = piVar1;
    if (*piVar1 + 1 != 0) {
      *piVar1 = *piVar1 + 1;
    }
    plVar8 = *(longlong **)(lVar7 + 0x28);
    pcVar12 = pcVar11;
    if (pcVar11 != _Py_NoneStruct_exref) goto LAB_1800c06da;
    pcVar12 = (code *)0x0;
  }
  *(code **)(lVar7 + 0x28) = pcVar12;
  if (((plVar8 != (longlong *)0x0) && (-1 < (int)*plVar8)) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)
     ) {
    (**(code **)(plVar8[1] + 0x30))();
  }
  if (((pcVar11 != (code *)0x0) && (-1 < *(int *)pcVar11)) &&
     (*(longlong *)pcVar11 = *(longlong *)pcVar11 + -1, *(longlong *)pcVar11 == 0)) {
    (**(code **)(*(longlong *)(pcVar11 + 8) + 0x30))(pcVar11);
  }
LAB_1800c0723:
  plVar4 = local_b0;
  plVar14 = local_f0;
  local_130 = *(undefined8 *)(local_100 + 0x108);
  local_138 = *(undefined8 *)(local_100 + 0xf0);
  local_140 = *(undefined8 *)(local_100 + 0x100);
  local_148 = *(undefined8 *)(local_100 + 0xe0);
  local_150 = *(undefined8 *)(local_100 + 0xf8);
  local_158 = *(undefined8 *)(local_100 + 0xe8);
  local_160 = local_d0;
  local_168 = local_d8;
  local_170 = local_e0;
  local_180 = local_f0;
  local_188 = local_b8;
  local_190 = local_c0;
  local_198 = local_c8;
  local_178 = plVar17;
  FUN_18036d0a0(plVar3,"oooooooooocccccc",plVar10,local_b0);
  plVar8 = DAT_1803b1ef0;
  if (plVar3 == DAT_1803b1ef0) {
    if ((-1 < (int)*DAT_1803b1ef0) && (*DAT_1803b1ef0 = *DAT_1803b1ef0 + -1, *plVar8 == 0)) {
      (**(code **)(plVar8[1] + 0x30))();
    }
    DAT_1803b1ef0 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if (((plVar4 != (longlong *)0x0) && (-1 < (int)*plVar4)) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)
     ) {
    (**(code **)(plVar4[1] + 0x30))(plVar4);
  }
  if (((local_c8 != (code *)0x0) && (-1 < *(int *)local_c8)) &&
     (*(longlong *)local_c8 = *(longlong *)local_c8 + -1, *(longlong *)local_c8 == 0)) {
    (**(code **)(*(longlong *)(local_c8 + 8) + 0x30))();
  }
  if (((local_c0 != (code *)0x0) && (-1 < *(int *)local_c0)) &&
     (*(longlong *)local_c0 = *(longlong *)local_c0 + -1, *(longlong *)local_c0 == 0)) {
    (**(code **)(*(longlong *)(local_c0 + 8) + 0x30))();
  }
  if (((local_b8 != (code *)0x0) && (-1 < *(int *)local_b8)) &&
     (*(longlong *)local_b8 = *(longlong *)local_b8 + -1, *(longlong *)local_b8 == 0)) {
    (**(code **)(*(longlong *)(local_b8 + 8) + 0x30))();
  }
  if (((plVar14 != (longlong *)0x0) && (-1 < (int)*plVar14)) &&
     (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
    (**(code **)(plVar14[1] + 0x30))(plVar14);
  }
  if (((plVar17 != (longlong *)0x0) && (-1 < (int)*plVar17)) &&
     (*plVar17 = *plVar17 + -1, *plVar17 == 0)) {
    (**(code **)(plVar17[1] + 0x30))(plVar17);
  }
  if (((local_e0 != (longlong *)0x0) && (-1 < (int)*local_e0)) &&
     (*local_e0 = *local_e0 + -1, *local_e0 == 0)) {
    (**(code **)(local_e0[1] + 0x30))();
  }
  if (((local_d8 != (longlong *)0x0) && (-1 < (int)*local_d8)) &&
     (*local_d8 = *local_d8 + -1, *local_d8 == 0)) {
    (**(code **)(local_d8[1] + 0x30))();
  }
  if (((local_d0 != (longlong *)0x0) && (-1 < (int)*local_d0)) &&
     (*local_d0 = *local_d0 + -1, *local_d0 == 0)) {
    (**(code **)(local_d0[1] + 0x30))();
  }
  if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
    (**(code **)(plVar10[1] + 0x30))(plVar10);
  }
  plVar10 = *(longlong **)(param_1 + 0x60);
  *(longlong *)(param_1 + 0x60) = lVar7;
  if (((plVar10 != (longlong *)0x0) && (-1 < (int)*plVar10)) &&
     (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
    (**(code **)(plVar10[1] + 0x30))();
  }
  return 0;
}
