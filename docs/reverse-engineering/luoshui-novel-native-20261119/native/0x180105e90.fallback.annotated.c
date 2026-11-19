/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x180105e90 */

/* WARNING: Function: __security_check_cookie replaced with injection: security_check_cookie */

undefined8 FUN_180105e90(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  code *pcVar1;
  undefined8 *puVar2;
  int *piVar3;
  bool bVar4;
  longlong *plVar5;
  char cVar6;
  int iVar7;
  code *pcVar8;
  longlong lVar9;
  longlong *plVar10;
  longlong *plVar11;
  longlong *plVar12;
  code *pcVar13;
  undefined8 uVar14;
  int iVar15;
  longlong lVar16;
  undefined8 uVar17;
  code *pcVar18;
  undefined1 auStack_198 [32];
  code *local_178;
  code *local_170;
  code *local_168;
  code *local_160;
  code *local_158;
  longlong *local_150;
  code *local_148;
  code *local_140;
  code *local_138;
  code *local_130;
  code *local_128;
  code *local_120;
  code *local_118;
  longlong *local_110;
  code *local_108;
  undefined8 local_100;
  code *local_f8;
  longlong *local_f0;
  code *local_e8;
  code *local_e0;
  code *local_d8;
  code *local_d0;
  code *local_c8;
  code *local_c0;
  longlong *local_b8;
  code *local_b0;
  code *local_a8;
  code *local_a0;
  longlong *local_98;
  code *local_90;
  code *local_88;
  code *local_80;
  code *local_78;
  code *local_70;
  code *local_68;
  code *local_60;
  code *local_58;
  longlong *local_50;
  ulonglong local_48;
  
  plVar10 = DAT_1803b3ab8;
  local_48 = DAT_1803a6040 ^ (ulonglong)auStack_198;
  local_a8 = (code *)param_3[2];
  plVar11 = (longlong *)*param_3;
  pcVar8 = (code *)param_3[1];
  local_f8 = (code *)param_3[3];
  local_e0 = (code *)param_3[4];
  local_d8 = (code *)param_3[5];
  local_a0 = (code *)param_3[6];
  local_98 = (longlong *)param_3[7];
  local_c0 = (code *)param_3[8];
  local_90 = (code *)param_3[9];
  local_c8 = (code *)param_3[10];
  local_d0 = (code *)0x0;
  pcVar18 = (code *)0x0;
  local_88 = (code *)0x0;
  local_80 = (code *)0x0;
  local_118 = (code *)0x0;
  local_f0 = DAT_180391d10;
  local_110 = DAT_1803b3ab8;
  local_b8 = plVar11;
  local_b0 = pcVar8;
  if (DAT_1803b3ab8 == (longlong *)0x0) {
LAB_180105f6d:
    DAT_1803b3ab8 = (longlong *)FUN_18036cd10(param_1,DAT_1803b3c10,DAT_1803ca778,0x70);
  }
  else if ((1 < *DAT_1803b3ab8) || (DAT_1803b3ab8[2] != 0)) {
    if ((DAT_1803b3ab8 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b3ab8 && (*DAT_1803b3ab8 = *DAT_1803b3ab8 + -1, *plVar10 == 0)))) {
      (**(code **)(plVar10[1] + 0x30))(plVar10);
    }
    goto LAB_180105f6d;
  }
  plVar10 = DAT_1803b3ab8;
  local_110 = DAT_1803b3ab8;
  FUN_180001d00(param_1,DAT_1803b3ab8);
  uVar17 = 0x38c;
  *(undefined4 *)(plVar10 + 5) = 0x38c;
/* novels.pipeline.pipeline_runner#22="_cancelled" */
  pcVar8 = (code *)FUN_18037e6c0(param_1,plVar11,DAT_1803b3fc0,pcVar8);
  if (pcVar8 == (code *)0x0) {
    plVar11 = *(longlong **)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
    plVar10 = local_110;
  }
  else if (pcVar8 == _Py_TrueStruct_exref) {
LAB_180105fde:
    bVar4 = true;
LAB_180106086:
    if ((-1 < *(int *)pcVar8) &&
       (*(longlong *)pcVar8 = *(longlong *)pcVar8 + -1, *(longlong *)pcVar8 == 0)) {
      (**(code **)(*(longlong *)(pcVar8 + 8) + 0x30))(pcVar8);
    }
    if (bVar4) {
      lVar9 = FUN_1800df4f0(param_1);
      if (lVar9 == 0) {
/* novels.pipeline.pipeline_runner#54="OperationCancelledError" */
        FUN_180373340(param_1,&local_f0,DAT_1803b40c0);
        uVar17 = 0x38d;
        plVar11 = local_f0;
        plVar10 = local_110;
      }
      else {
        uVar17 = 0x38d;
        *(undefined4 *)(plVar10 + 5) = 0x38d;
        plVar10 = (longlong *)FUN_1803770d0(param_1,lVar9,DAT_1803b40c8);
        if (plVar10 == (longlong *)0x0) {
          plVar11 = *(longlong **)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          plVar10 = local_110;
        }
        else {
          plVar11 = (longlong *)FUN_180373490(param_1,plVar10);
          if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
            (**(code **)(plVar10[1] + 0x30))(plVar10);
          }
          if (plVar11 == (longlong *)0x0) {
            plVar11 = *(longlong **)(param_1 + 0x60);
            *(undefined8 *)(param_1 + 0x60) = 0;
            plVar10 = local_110;
          }
          else {
            FUN_18000d1d0(param_1,plVar11);
            plVar10 = local_110;
          }
        }
      }
    }
    else {
      plVar12 = (longlong *)FUN_18037f0c0();
      pcVar8 = local_b0;
      uVar17 = 0x38e;
      if (plVar12 == (longlong *)0x0) {
        plVar11 = *(longlong **)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
        plVar10 = local_110;
      }
      else {
        *(undefined4 *)(plVar10 + 5) = 0x38e;
/* novels.pipeline.pipeline_runner#87="story_dir" */
        local_e8 = (code *)FUN_18037e6c0(param_1,plVar12,DAT_1803b41c8,local_b0);
        if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
          (**(code **)(plVar12[1] + 0x30))(plVar12);
        }
/* novels.pipeline.pipeline_runner#420="生成最终真相文件" */
        pcVar1 = DAT_1803b4c30;
/* novels.pipeline.pipeline_runner#90="active" */
        pcVar13 = DAT_1803b41e0;
        if (local_e8 == (code *)0x0) {
          plVar11 = *(longlong **)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          plVar10 = local_110;
        }
        else {
          uVar17 = 0x38f;
          *(undefined4 *)(plVar10 + 5) = 0x38f;
          local_78 = local_c0;
          local_68 = pcVar1;
          local_70 = pcVar13;
          local_d0 = local_e8;
/* novels.pipeline.pipeline_runner#42="_stage" */
          plVar12 = (longlong *)FUN_18037e980(param_1,plVar11,DAT_1803b4060,&local_78);
          if (plVar12 == (longlong *)0x0) {
            plVar11 = *(longlong **)(param_1 + 0x60);
            *(undefined8 *)(param_1 + 0x60) = 0;
            plVar10 = local_110;
          }
          else {
            if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
              (**(code **)(plVar12[1] + 0x30))();
            }
            uVar17 = 0x390;
            *(undefined4 *)(plVar10 + 5) = 0x390;
            plVar12 = (longlong *)
/* novels.pipeline.pipeline_runner#106="_log" */
                      FUN_18037e6c0(param_1,plVar11,DAT_1803b4260,
                                    *(undefined8 *)(DAT_1803b4c38 + 0x18));
            if (plVar12 == (longlong *)0x0) {
              plVar11 = *(longlong **)(param_1 + 0x60);
              *(undefined8 *)(param_1 + 0x60) = 0;
              plVar10 = local_110;
            }
            else {
              if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                (**(code **)(plVar12[1] + 0x30))();
              }
              local_70 = local_a8;
              uVar17 = 0x391;
              local_68 = local_f8;
              local_60 = local_e0;
              local_58 = local_d8;
              *(undefined4 *)(plVar10 + 5) = 0x391;
              local_50 = local_98;
              local_78 = pcVar8;
/* novels.pipeline.pipeline_runner#355="_settle_validate_write_truth" */
              plVar11 = (longlong *)FUN_18037eda0(param_1,plVar11,DAT_1803b4a28,&local_78);
              if (plVar11 == (longlong *)0x0) {
                plVar11 = *(longlong **)(param_1 + 0x60);
                *(undefined8 *)(param_1 + 0x60) = 0;
                plVar10 = local_110;
              }
              else {
                plVar12 = (longlong *)FUN_180001a60(plVar11);
                if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                  (**(code **)(plVar11[1] + 0x30))(plVar11);
                }
                if (plVar12 == (longlong *)0x0) {
                  plVar11 = *(longlong **)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                  plVar10 = local_110;
                }
                else {
                  pcVar13 = (code *)(**(code **)(plVar12[1] + 0xe0))(plVar12);
                  if (pcVar13 == (code *)0x0) {
                    plVar11 = *(longlong **)(param_1 + 0x60);
                    if ((plVar11 != (longlong *)0x0) && (lVar9 = plVar11[1], lVar9 != 0)) {
                      lVar16 = *(longlong *)(lVar9 + 8);
                      if ((*(uint *)(*(longlong *)(lVar9 + 8) + 0xa8) & 0x40000000) == 0) {
                        lVar16 = lVar9;
                      }
                      if ((lVar16 != *(longlong *)PyExc_StopIteration_exref) &&
                         (((-1 < *(int *)(*(longlong *)(lVar16 + 8) + 0xa8) ||
                           ((*(uint *)(lVar16 + 0xa8) & 0x40000000) == 0)) ||
                          (cVar6 = FUN_180382640(lVar16,*(longlong *)PyExc_StopIteration_exref),
                          cVar6 != '\x01')))) {
                        *(undefined8 *)(param_1 + 0x60) = 0;
                        pcVar13 = (code *)0x0;
                        goto LAB_180106e31;
                      }
                      *(undefined8 *)(param_1 + 0x60) = 0;
                      if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                        (**(code **)(plVar11[1] + 0x30))(plVar11);
                      }
                    }
                    plVar10 = (longlong *)
                              PyUnicode_FromFormat
                                        ("not enough values to unpack (expected %d, got %d)",2,0);
                    plVar11 = (longlong *)
                              FUN_180376770(param_1,*(undefined8 *)PyExc_ValueError_exref,plVar10);
                    pcVar13 = pcVar18;
                    if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                      (**(code **)(plVar10[1] + 0x30))(plVar10);
                    }
                  }
                  else {
                    pcVar18 = (code *)(**(code **)(plVar12[1] + 0xe0))(plVar12);
                    if (pcVar18 == (code *)0x0) {
                      plVar11 = *(longlong **)(param_1 + 0x60);
                      if ((plVar11 != (longlong *)0x0) && (lVar9 = plVar11[1], lVar9 != 0)) {
                        lVar16 = *(longlong *)(lVar9 + 8);
                        if ((*(uint *)(*(longlong *)(lVar9 + 8) + 0xa8) & 0x40000000) == 0) {
                          lVar16 = lVar9;
                        }
                        if ((lVar16 != *(longlong *)PyExc_StopIteration_exref) &&
                           (((-1 < *(int *)(*(longlong *)(lVar16 + 8) + 0xa8) ||
                             ((*(uint *)(lVar16 + 0xa8) & 0x40000000) == 0)) ||
                            (cVar6 = FUN_180382640(lVar16,*(longlong *)PyExc_StopIteration_exref),
                            cVar6 != '\x01')))) {
LAB_180106473:
                          *(undefined8 *)(param_1 + 0x60) = 0;
                          goto LAB_180106e31;
                        }
                        *(undefined8 *)(param_1 + 0x60) = 0;
                        if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                          (**(code **)(plVar11[1] + 0x30))(plVar11);
                        }
                      }
                      plVar10 = (longlong *)
                                PyUnicode_FromFormat
                                          ("not enough values to unpack (expected %d, got %d)",2,1);
                    }
                    else {
                      local_118 = pcVar18;
                      plVar11 = (longlong *)(**(code **)(plVar12[1] + 0xe0))(plVar12);
                      if (plVar11 == (longlong *)0x0) {
                        plVar11 = *(longlong **)(param_1 + 0x60);
                        if ((plVar11 != (longlong *)0x0) && (lVar9 = plVar11[1], lVar9 != 0)) {
                          lVar16 = *(longlong *)(lVar9 + 8);
                          if ((*(uint *)(*(longlong *)(lVar9 + 8) + 0xa8) & 0x40000000) == 0) {
                            lVar16 = lVar9;
                          }
                          if (lVar16 == *(longlong *)PyExc_StopIteration_exref) {
                            *(undefined8 *)(param_1 + 0x60) = 0;
                          }
                          else {
                            if ((-1 < *(int *)(*(longlong *)(lVar16 + 8) + 0xa8)) ||
                               ((*(uint *)(lVar16 + 0xa8) & 0x40000000) == 0)) goto LAB_180106473;
                            cVar6 = FUN_180382640(lVar16,*(longlong *)PyExc_StopIteration_exref);
                            *(undefined8 *)(param_1 + 0x60) = 0;
                            if (cVar6 != '\x01') goto LAB_180106e31;
                          }
                          if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                            (**(code **)(plVar11[1] + 0x30))(plVar11);
                          }
                        }
                        if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                          (**(code **)(plVar12[1] + 0x30))(plVar12);
                        }
                        iVar7 = *(int *)pcVar13 + 1;
                        iVar15 = *(int *)pcVar13;
                        if (iVar7 != 0) {
                          *(int *)pcVar13 = iVar7;
                          iVar15 = iVar7;
                        }
                        local_88 = pcVar13;
                        if ((-1 < iVar15) &&
                           (*(longlong *)pcVar13 = *(longlong *)pcVar13 + -1,
                           *(longlong *)pcVar13 == 0)) {
                          (**(code **)(*(longlong *)(pcVar13 + 8) + 0x30))(pcVar13);
                        }
                        iVar7 = *(int *)pcVar18 + 1;
                        iVar15 = *(int *)pcVar18;
                        if (iVar7 != 0) {
                          *(int *)pcVar18 = iVar7;
                          iVar15 = iVar7;
                        }
                        local_80 = pcVar18;
                        if ((-1 < iVar15) &&
                           (*(longlong *)pcVar18 = *(longlong *)pcVar18 + -1,
                           *(longlong *)pcVar18 == 0)) {
                          (**(code **)(*(longlong *)(pcVar18 + 8) + 0x30))(pcVar18);
                        }
                        plVar11 = local_b8;
/* novels.pipeline.pipeline_runner#282="completed" */
                        local_100 = DAT_1803b47e0;
                        uVar17 = 0x394;
                        *(undefined4 *)(plVar10 + 5) = 0x394;
                        local_108 = local_c0;
                        plVar12 = (longlong *)
/* novels.pipeline.pipeline_runner#42="_stage" */
                                  FUN_18037e820(param_1,local_b8,DAT_1803b4060,&local_108);
                        if (plVar12 == (longlong *)0x0) {
                          plVar11 = *(longlong **)(param_1 + 0x60);
                          *(undefined8 *)(param_1 + 0x60) = 0;
                          plVar10 = local_110;
                        }
                        else {
                          if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                            (**(code **)(plVar12[1] + 0x30))();
                          }
                          pcVar8 = local_90;
/* novels.pipeline.pipeline_runner#422="校验真相文件变更" */
                          local_68 = DAT_1803b4c40;
/* novels.pipeline.pipeline_runner#90="active" */
                          local_70 = DAT_1803b41e0;
                          uVar17 = 0x395;
                          *(undefined4 *)(plVar10 + 5) = 0x395;
                          local_78 = local_90;
                          plVar12 = (longlong *)
/* novels.pipeline.pipeline_runner#42="_stage" */
                                    FUN_18037e980(param_1,plVar11,DAT_1803b4060,&local_78);
                          if (plVar12 == (longlong *)0x0) {
                            plVar11 = *(longlong **)(param_1 + 0x60);
                            *(undefined8 *)(param_1 + 0x60) = 0;
                            plVar10 = local_110;
                          }
                          else {
                            if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                              (**(code **)(plVar12[1] + 0x30))();
                            }
/* novels.pipeline.pipeline_runner#282="completed" */
                            local_100 = DAT_1803b47e0;
                            uVar17 = 0x396;
                            local_108 = pcVar8;
                            *(undefined4 *)(plVar10 + 5) = 0x396;
                            plVar12 = (longlong *)
/* novels.pipeline.pipeline_runner#42="_stage" */
                                      FUN_18037e820(param_1,plVar11,DAT_1803b4060,&local_108);
                            if (plVar12 == (longlong *)0x0) {
                              plVar11 = *(longlong **)(param_1 + 0x60);
                              *(undefined8 *)(param_1 + 0x60) = 0;
                              plVar10 = local_110;
                            }
                            else {
                              if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0))
                              {
                                (**(code **)(plVar12[1] + 0x30))();
                              }
/* novels.pipeline.pipeline_runner#423="同步记忆索引" */
                              local_68 = DAT_1803b4c48;
/* novels.pipeline.pipeline_runner#90="active" */
                              local_70 = DAT_1803b41e0;
                              uVar17 = 0x397;
                              *(undefined4 *)(plVar10 + 5) = 0x397;
                              local_78 = local_c8;
                              plVar11 = (longlong *)
/* novels.pipeline.pipeline_runner#42="_stage" */
                                        FUN_18037e980(param_1,plVar11,DAT_1803b4060,&local_78);
                              if (plVar11 == (longlong *)0x0) {
                                plVar11 = *(longlong **)(param_1 + 0x60);
                                *(undefined8 *)(param_1 + 0x60) = 0;
                                plVar10 = local_110;
                              }
                              else {
                                if ((-1 < (int)*plVar11) &&
                                   (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                  (**(code **)(plVar11[1] + 0x30))();
                                }
                                lVar9 = FUN_1800e0f10(param_1);
                                if (lVar9 == 0) {
/* novels.pipeline.pipeline_runner#356="upsert_chapter_entry" */
                                  FUN_180373340(param_1,&local_f0,DAT_1803b4a30);
                                  uVar17 = 0x398;
                                  plVar11 = local_f0;
                                  plVar10 = local_110;
                                }
                                else {
                                  plVar10 = (longlong *)FUN_18037f0c0();
                                  if (plVar10 == (longlong *)0x0) {
                                    plVar11 = *(longlong **)(param_1 + 0x60);
                                    uVar17 = 0x39d;
                                    *(undefined8 *)(param_1 + 0x60) = 0;
                                    plVar10 = local_110;
                                  }
                                  else {
                                    iVar7 = FUN_180001510(plVar10);
                                    if (iVar7 == -1) {
                                      plVar11 = *(longlong **)(param_1 + 0x60);
                                      *(undefined8 *)(param_1 + 0x60) = 0;
                                      if ((-1 < (int)*plVar10) &&
                                         (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                        (**(code **)(plVar10[1] + 0x30))(plVar10);
                                      }
                                      uVar17 = 0x39d;
                                      plVar10 = local_110;
                                    }
                                    else {
                                      if (iVar7 != 1) {
                                        if ((-1 < (int)*plVar10) &&
                                           (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                          (**(code **)(plVar10[1] + 0x30))(plVar10);
                                        }
/* novels.pipeline.pipeline_runner#4="" */
                                        plVar10 = DAT_1803b3f30;
/* novels.pipeline.pipeline_runner#4="" */
                                        iVar7 = (int)*DAT_1803b3f30 + 1;
                                        if (iVar7 != 0) {
/* novels.pipeline.pipeline_runner#4="" */
                                          *(int *)DAT_1803b3f30 = iVar7;
                                        }
                                      }
                                      plVar11 = (longlong *)FUN_18037f0c0();
                                      if ((-1 < (int)*plVar10) &&
                                         (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                        (**(code **)(plVar10[1] + 0x30))(plVar10);
                                      }
                                      uVar17 = 0x39d;
                                      if (plVar11 == (longlong *)0x0) {
                                        plVar11 = *(longlong **)(param_1 + 0x60);
                                        *(undefined8 *)(param_1 + 0x60) = 0;
                                        plVar10 = local_110;
                                      }
                                      else {
                                        *(undefined4 *)(local_110 + 5) = 0x39d;
                                        pcVar8 = (code *)FUN_180376210(param_1,plVar11);
                                        if ((-1 < (int)*plVar11) &&
                                           (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                          (**(code **)(plVar11[1] + 0x30))(plVar11);
                                        }
                                        if (pcVar8 == (code *)0x0) {
                                          plVar11 = *(longlong **)(param_1 + 0x60);
                                          *(undefined8 *)(param_1 + 0x60) = 0;
                                          plVar10 = local_110;
                                        }
                                        else {
                                          uVar17 = 0x398;
                                          *(undefined4 *)(local_110 + 5) = 0x398;
                                          local_118 = local_e8;
                                          local_78 = local_f8;
                                          local_70 = local_e0;
                                          local_68 = local_d8;
                                          local_58 = local_a0;
                                          local_178 = (code *)DAT_1803b4a40;
                                          local_60 = pcVar8;
                                          plVar11 = (longlong *)
                                                    FUN_18037d030(param_1,lVar9,&local_118,&local_78
                                                                 );
                                          if ((-1 < *(int *)pcVar8) &&
                                             (*(longlong *)pcVar8 = *(longlong *)pcVar8 + -1,
                                             *(longlong *)pcVar8 == 0)) {
                                            (**(code **)(*(longlong *)(pcVar8 + 8) + 0x30))(pcVar8);
                                          }
                                          if (plVar11 == (longlong *)0x0) {
                                            plVar11 = *(longlong **)(param_1 + 0x60);
                                            *(undefined8 *)(param_1 + 0x60) = 0;
                                            plVar10 = local_110;
                                          }
                                          else {
                                            if ((-1 < (int)*plVar11) &&
                                               (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                              (**(code **)(plVar11[1] + 0x30))(plVar11);
                                            }
                                            plVar10 = (longlong *)FUN_18037f0c0();
/* novels.pipeline.pipeline_runner#424="记忆索引已更新：第 " */
                                            piVar3 = DAT_1803b4c50;
                                            if (plVar10 == (longlong *)0x0) {
                                              plVar11 = *(longlong **)(param_1 + 0x60);
                                              uVar17 = 0x3a0;
                                              *(undefined8 *)(param_1 + 0x60) = 0;
                                              plVar10 = local_110;
                                            }
                                            else {
                                              lVar9 = *(longlong *)(param_1 + 0x10);
                                              pcVar8 = *(code **)(lVar9 + 0x41da8);
                                              if (pcVar8 == (code *)0x0) {
                                                pcVar8 = (code *)FUN_18035ebd0(PyTuple_Type_exref,3)
                                                ;
                                              }
                                              else {
                                                *(longlong *)(lVar9 + 0x41da8) =
                                                     *(longlong *)(pcVar8 + 0x18);
                                                *(int *)(lVar9 + 0x41e40) =
                                                     *(int *)(lVar9 + 0x41e40) + -1;
                                                *(longlong *)pcVar8 = 1;
                                              }
                                              pcVar1 = pcVar8 + -0x10;
                                              *(longlong *)(pcVar8 + 0x18) = 0;
                                              *(longlong *)(pcVar8 + 0x20) = 0;
                                              *(longlong *)(pcVar8 + 0x28) = 0;
                                              lVar9 = _PyThreadState_GetCurrent();
                                              lVar9 = *(longlong *)
                                                       (*(longlong *)(lVar9 + 0x10) + 0xd0);
                                              puVar2 = *(undefined8 **)(lVar9 + 8);
                                              *puVar2 = pcVar1;
                                              *(ulonglong *)(pcVar8 + -8) =
                                                   *(ulonglong *)(pcVar8 + -8) & 3;
                                              *(ulonglong *)(pcVar8 + -8) =
                                                   *(ulonglong *)(pcVar8 + -8) | (ulonglong)puVar2;
                                              *(longlong *)pcVar1 = lVar9;
                                              *(code **)(lVar9 + 8) = pcVar1;
                                              if (*piVar3 + 1 != 0) {
                                                *piVar3 = *piVar3 + 1;
                                              }
                                              *(int **)(pcVar8 + 0x18) = piVar3;
/* novels.pipeline.pipeline_runner#4="" */
                                              lVar9 = FUN_180375520(param_1,local_f8,DAT_1803b3f30);
                                              if (lVar9 == 0) {
                                                plVar11 = *(longlong **)(param_1 + 0x60);
                                                uVar17 = 0x3a0;
                                                *(undefined8 *)(param_1 + 0x60) = 0;
                                                if ((-1 < (int)*plVar10) &&
                                                   (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                                  (**(code **)(plVar10[1] + 0x30))(plVar10);
                                                }
                                                plVar10 = local_110;
                                                if ((-1 < *(int *)pcVar8) &&
                                                   (*(longlong *)pcVar8 = *(longlong *)pcVar8 + -1,
                                                   *(longlong *)pcVar8 == 0)) {
                                                  lVar9 = *(longlong *)(pcVar8 + 8);
                                                  goto LAB_180106e82;
                                                }
                                              }
                                              else {
                                                *(longlong *)(pcVar8 + 0x20) = lVar9;
/* novels.pipeline.pipeline_runner#371=" 章" */
                                                piVar3 = DAT_1803b4aa8;
/* novels.pipeline.pipeline_runner#371=" 章" */
                                                if (*DAT_1803b4aa8 + 1 != 0) {
/* novels.pipeline.pipeline_runner#371=" 章" */
                                                  *DAT_1803b4aa8 = *DAT_1803b4aa8 + 1;
                                                }
                                                *(int **)(pcVar8 + 0x28) = piVar3;
                                                plVar11 = (longlong *)
/* novels.pipeline.pipeline_runner#4="" */
                                                          PyUnicode_Join(DAT_1803b3f30,pcVar8);
                                                if ((-1 < *(int *)pcVar8) &&
                                                   (*(longlong *)pcVar8 = *(longlong *)pcVar8 + -1,
                                                   *(longlong *)pcVar8 == 0)) {
                                                  (**(code **)(*(longlong *)(pcVar8 + 8) + 0x30))
                                                            (pcVar8);
                                                }
                                                if (plVar11 == (longlong *)0x0) {
                                                  plVar11 = *(longlong **)(param_1 + 0x60);
                                                  *(undefined8 *)(param_1 + 0x60) = 0;
                                                  if ((-1 < (int)*plVar10) &&
                                                     (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                                    (**(code **)(plVar10[1] + 0x30))(plVar10);
                                                  }
                                                  uVar17 = 0x3a0;
                                                  plVar10 = local_110;
                                                }
                                                else {
                                                  uVar17 = 0x3a0;
                                                  *(undefined4 *)(local_110 + 5) = 0x3a0;
                                                  plVar12 = (longlong *)
                                                            FUN_180376770(param_1,plVar10,plVar11);
                                                  if ((-1 < (int)*plVar10) &&
                                                     (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                                    (**(code **)(plVar10[1] + 0x30))(plVar10);
                                                  }
                                                  if ((-1 < (int)*plVar11) &&
                                                     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                    (**(code **)(plVar11[1] + 0x30))(plVar11);
                                                  }
                                                  if (plVar12 == (longlong *)0x0) {
                                                    plVar11 = *(longlong **)(param_1 + 0x60);
                                                    *(undefined8 *)(param_1 + 0x60) = 0;
                                                    plVar10 = local_110;
                                                  }
                                                  else {
                                                    if ((-1 < (int)*plVar12) &&
                                                       (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                                      (**(code **)(plVar12[1] + 0x30))(plVar12);
                                                    }
                                                    plVar11 = local_b8;
                                                    plVar10 = local_110;
                                                    uVar17 = 0x3a1;
                                                    local_78 = local_d0;
                                                    local_70 = local_f8;
                                                    local_68 = local_e0;
                                                    *(undefined4 *)(local_110 + 5) = 0x3a1;
                                                    local_60 = local_d8;
                                                    local_58 = pcVar13;
                                                    plVar12 = (longlong *)
                                                              FUN_18037ec40(param_1,local_b8,
/* novels.pipeline.pipeline_runner#425="_sync_temporal_memory" */
                                                                            DAT_1803b4c58,&local_78)
                                                    ;
                                                    if (plVar12 == (longlong *)0x0) {
                                                      plVar11 = *(longlong **)(param_1 + 0x60);
                                                      *(undefined8 *)(param_1 + 0x60) = 0;
                                                    }
                                                    else {
                                                      if ((-1 < (int)*plVar12) &&
                                                         (*plVar12 = *plVar12 + -1, *plVar12 == 0))
                                                      {
                                                        (**(code **)(plVar12[1] + 0x30))(plVar12);
                                                      }
/* novels.pipeline.pipeline_runner#282="completed" */
                                                      local_100 = DAT_1803b47e0;
                                                      uVar17 = 0x3a4;
                                                      *(undefined4 *)(plVar10 + 5) = 0x3a4;
                                                      local_108 = local_c8;
                                                      plVar12 = (longlong *)
                                                                FUN_18037e820(param_1,plVar11,
/* novels.pipeline.pipeline_runner#42="_stage" */
                                                                              DAT_1803b4060,
                                                                              &local_108);
                                                      if (plVar12 != (longlong *)0x0) {
                                                        if ((-1 < (int)*plVar12) &&
                                                           (*plVar12 = *plVar12 + -1, *plVar12 == 0)
                                                           ) {
                                                          (**(code **)(plVar12[1] + 0x30))(plVar12);
                                                        }
                                                        FUN_180001d80(param_1);
/* novels.pipeline.pipeline_runner#206="settlement" */
                                                        uVar17 = DAT_1803b4580;
                                                        uVar14 = _PyDict_NewPresized(2);
                                                        PyDict_SetItem(uVar14,uVar17,pcVar13);
/* novels.pipeline.pipeline_runner#205="validation" */
                                                        PyDict_SetItem(uVar14,DAT_1803b4578,pcVar18)
                                                        ;
                                                        if ((-1 < *(int *)local_e8) &&
                                                           (*(longlong *)local_e8 =
                                                                 *(longlong *)local_e8 + -1,
                                                           *(longlong *)local_e8 == 0)) {
                                                          (**(code **)(*(longlong *)(local_d0 + 8) +
                                                                      0x30))();
                                                        }
                                                        if ((-1 < *(int *)pcVar13) &&
                                                           (*(longlong *)pcVar13 =
                                                                 *(longlong *)pcVar13 + -1,
                                                           *(longlong *)pcVar13 == 0)) {
                                                          (**(code **)(*(longlong *)(pcVar13 + 8) +
                                                                      0x30))(pcVar13);
                                                        }
                                                        if ((-1 < *(int *)pcVar18) &&
                                                           (*(longlong *)pcVar18 =
                                                                 *(longlong *)pcVar18 + -1,
                                                           *(longlong *)pcVar18 == 0)) {
                                                          (**(code **)(*(longlong *)(pcVar18 + 8) +
                                                                      0x30))(pcVar18);
                                                        }
                                                        if ((-1 < (int)*plVar11) &&
                                                           (*plVar11 = *plVar11 + -1, *plVar11 == 0)
                                                           ) {
                                                          (**(code **)(plVar11[1] + 0x30))(plVar11);
                                                        }
                                                        if ((-1 < *(int *)local_b0) &&
                                                           (*(longlong *)local_b0 =
                                                                 *(longlong *)local_b0 + -1,
                                                           *(longlong *)local_b0 == 0)) {
                                                          (**(code **)(*(longlong *)(local_b0 + 8) +
                                                                      0x30))();
                                                        }
                                                        if ((-1 < *(int *)local_a8) &&
                                                           (*(longlong *)local_a8 =
                                                                 *(longlong *)local_a8 + -1,
                                                           *(longlong *)local_a8 == 0)) {
                                                          (**(code **)(*(longlong *)(local_a8 + 8) +
                                                                      0x30))();
                                                        }
                                                        if ((-1 < *(int *)local_f8) &&
                                                           (*(longlong *)local_f8 =
                                                                 *(longlong *)local_f8 + -1,
                                                           *(longlong *)local_f8 == 0)) {
                                                          (**(code **)(*(longlong *)(local_f8 + 8) +
                                                                      0x30))();
                                                        }
                                                        if ((-1 < *(int *)local_e0) &&
                                                           (*(longlong *)local_e0 =
                                                                 *(longlong *)local_e0 + -1,
                                                           *(longlong *)local_e0 == 0)) {
                                                          (**(code **)(*(longlong *)(local_e0 + 8) +
                                                                      0x30))();
                                                        }
                                                        if ((-1 < *(int *)local_d8) &&
                                                           (*(longlong *)local_d8 =
                                                                 *(longlong *)local_d8 + -1,
                                                           *(longlong *)local_d8 == 0)) {
                                                          (**(code **)(*(longlong *)(local_d8 + 8) +
                                                                      0x30))();
                                                        }
                                                        if ((-1 < *(int *)local_a0) &&
                                                           (*(longlong *)local_a0 =
                                                                 *(longlong *)local_a0 + -1,
                                                           *(longlong *)local_a0 == 0)) {
                                                          (**(code **)(*(longlong *)(local_a0 + 8) +
                                                                      0x30))();
                                                        }
                                                        if ((-1 < (int)*local_98) &&
                                                           (*local_98 = *local_98 + -1,
                                                           *local_98 == 0)) {
                                                          (**(code **)(local_98[1] + 0x30))();
                                                        }
                                                        if ((-1 < *(int *)local_c0) &&
                                                           (*(longlong *)local_c0 =
                                                                 *(longlong *)local_c0 + -1,
                                                           *(longlong *)local_c0 == 0)) {
                                                          (**(code **)(*(longlong *)(local_c0 + 8) +
                                                                      0x30))();
                                                        }
                                                        if ((-1 < *(int *)local_90) &&
                                                           (*(longlong *)local_90 =
                                                                 *(longlong *)local_90 + -1,
                                                           *(longlong *)local_90 == 0)) {
                                                          (**(code **)(*(longlong *)(local_90 + 8) +
                                                                      0x30))();
                                                        }
                                                        if (*(int *)local_c8 < 0) {
                                                          return uVar14;
                                                        }
                                                        *(longlong *)local_c8 =
                                                             *(longlong *)local_c8 + -1;
                                                        if (*(longlong *)local_c8 != 0) {
                                                          return uVar14;
                                                        }
                                                        (**(code **)(*(longlong *)(local_c8 + 8) +
                                                                    0x30))();
                                                        return uVar14;
                                                      }
                                                      plVar11 = *(longlong **)(param_1 + 0x60);
                                                      *(undefined8 *)(param_1 + 0x60) = 0;
                                                    }
                                                  }
                                                }
                                              }
                                            }
                                          }
                                        }
                                      }
                                    }
                                  }
                                }
                              }
                            }
                          }
                        }
                        goto LAB_180106e8c;
                      }
                      if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                        (**(code **)(plVar11[1] + 0x30))(plVar11);
                      }
                      plVar10 = (longlong *)
                                PyUnicode_FromFormat("too many values to unpack (expected %d)",2);
                    }
                    plVar11 = (longlong *)
                              FUN_180376770(param_1,*(undefined8 *)PyExc_ValueError_exref,plVar10);
                    if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                      (**(code **)(plVar10[1] + 0x30))(plVar10);
                    }
                  }
LAB_180106e31:
                  if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                    (**(code **)(plVar12[1] + 0x30))(plVar12);
                  }
                  if (((pcVar13 != (code *)0x0) && (-1 < *(int *)pcVar13)) &&
                     (*(longlong *)pcVar13 = *(longlong *)pcVar13 + -1, *(longlong *)pcVar13 == 0))
                  {
                    (**(code **)(*(longlong *)(pcVar13 + 8) + 0x30))(pcVar13);
                  }
                  plVar10 = local_110;
                  if (((local_118 != (code *)0x0) && (-1 < *(int *)local_118)) &&
                     (*(longlong *)local_118 = *(longlong *)local_118 + -1,
                     *(longlong *)local_118 == 0)) {
                    lVar9 = *(longlong *)(local_118 + 8);
                    pcVar8 = local_118;
                    goto LAB_180106e82;
                  }
                }
              }
            }
          }
        }
      }
    }
  }
  else {
    if ((pcVar8 == _Py_FalseStruct_exref) || (pcVar8 == _Py_NoneStruct_exref)) {
LAB_180106084:
      bVar4 = false;
      goto LAB_180106086;
    }
    lVar9 = *(longlong *)(pcVar8 + 8);
    if ((*(longlong *)(lVar9 + 0x60) == 0) ||
       (pcVar13 = *(code **)(*(longlong *)(lVar9 + 0x60) + 0x48), pcVar13 == (code *)0x0)) {
      if ((*(undefined8 **)(lVar9 + 0x70) == (undefined8 *)0x0) ||
         (pcVar13 = (code *)**(undefined8 **)(lVar9 + 0x70), pcVar13 == (code *)0x0)) {
        if ((*(undefined8 **)(lVar9 + 0x68) == (undefined8 *)0x0) ||
           (pcVar13 = (code *)**(undefined8 **)(lVar9 + 0x68), pcVar13 == (code *)0x0))
        goto LAB_180105fde;
        lVar9 = (*pcVar13)(pcVar8);
      }
      else {
        lVar9 = (*pcVar13)(pcVar8);
      }
    }
    else {
      iVar7 = (*pcVar13)(pcVar8);
      lVar9 = (longlong)iVar7;
    }
    if (0 < lVar9) {
      bVar4 = true;
      goto LAB_180106086;
    }
    if (lVar9 == 0) goto LAB_180106084;
    plVar11 = *(longlong **)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
    plVar10 = local_110;
    if ((*(int *)pcVar8 < 0) ||
       (*(longlong *)pcVar8 = *(longlong *)pcVar8 + -1, *(longlong *)pcVar8 != 0))
    goto LAB_180106e8c;
    lVar9 = *(longlong *)(pcVar8 + 8);
LAB_180106e82:
    (**(code **)(lVar9 + 0x30))(pcVar8);
    plVar10 = local_110;
  }
LAB_180106e8c:
  piVar3 = (int *)plVar11[5];
  if (piVar3 == (int *)0x0) {
    pcVar8 = (code *)FUN_18038e2a0(plVar10,uVar17);
    plVar12 = (longlong *)plVar11[5];
    pcVar18 = pcVar8;
    if (pcVar8 == _Py_NoneStruct_exref) {
      pcVar18 = (code *)0x0;
    }
    else {
LAB_180106ee6:
      pcVar8 = pcVar18;
      if ((pcVar18 != (code *)0x0) && (*(int *)pcVar18 + 1 != 0)) {
        *(int *)pcVar18 = *(int *)pcVar18 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar3 + 6) == plVar10) goto LAB_180106f2f;
    pcVar8 = (code *)FUN_18038e2a0(plVar10,uVar17);
    *(int **)(pcVar8 + 0x10) = piVar3;
    if (*piVar3 + 1 != 0) {
      *piVar3 = *piVar3 + 1;
    }
    plVar12 = (longlong *)plVar11[5];
    pcVar18 = pcVar8;
    if (pcVar8 != _Py_NoneStruct_exref) goto LAB_180106ee6;
    pcVar18 = (code *)0x0;
  }
  plVar11[5] = (longlong)pcVar18;
  if (((plVar12 != (longlong *)0x0) && (-1 < (int)*plVar12)) &&
     (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
    (**(code **)(plVar12[1] + 0x30))();
  }
  if (((pcVar8 != (code *)0x0) && (-1 < *(int *)pcVar8)) &&
     (*(longlong *)pcVar8 = *(longlong *)pcVar8 + -1, *(longlong *)pcVar8 == 0)) {
    (**(code **)(*(longlong *)(pcVar8 + 8) + 0x30))(pcVar8);
  }
LAB_180106f2f:
  pcVar13 = local_90;
  plVar5 = local_b8;
  pcVar18 = local_c0;
  pcVar8 = local_c8;
  local_120 = local_80;
  local_128 = local_88;
  local_130 = local_d0;
  local_138 = local_c8;
  local_140 = local_90;
  local_148 = local_c0;
  local_150 = local_98;
  local_158 = local_a0;
  local_160 = local_d8;
  local_168 = local_e0;
  local_170 = local_f8;
  local_178 = local_a8;
  FUN_18036d0a0(plVar10,"oooooooooooooo",local_b8,local_b0);
  plVar12 = DAT_1803b3ab8;
  if (plVar10 == DAT_1803b3ab8) {
    if ((-1 < (int)*DAT_1803b3ab8) && (*DAT_1803b3ab8 = *DAT_1803b3ab8 + -1, *plVar12 == 0)) {
      (**(code **)(plVar12[1] + 0x30))();
    }
    DAT_1803b3ab8 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if (((local_d0 != (code *)0x0) && (-1 < *(int *)local_d0)) &&
     (*(longlong *)local_d0 = *(longlong *)local_d0 + -1, *(longlong *)local_d0 == 0)) {
    (**(code **)(*(longlong *)(local_d0 + 8) + 0x30))();
  }
  if (((local_88 != (code *)0x0) && (-1 < *(int *)local_88)) &&
     (*(longlong *)local_88 = *(longlong *)local_88 + -1, *(longlong *)local_88 == 0)) {
    (**(code **)(*(longlong *)(local_88 + 8) + 0x30))();
  }
  if (((local_80 != (code *)0x0) && (-1 < *(int *)local_80)) &&
     (*(longlong *)local_80 = *(longlong *)local_80 + -1, *(longlong *)local_80 == 0)) {
    (**(code **)(*(longlong *)(local_80 + 8) + 0x30))();
  }
  if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
    (**(code **)(plVar5[1] + 0x30))(plVar5);
  }
  if ((-1 < *(int *)local_b0) &&
     (*(longlong *)local_b0 = *(longlong *)local_b0 + -1, *(longlong *)local_b0 == 0)) {
    (**(code **)(*(longlong *)(local_b0 + 8) + 0x30))();
  }
  if ((-1 < *(int *)local_a8) &&
     (*(longlong *)local_a8 = *(longlong *)local_a8 + -1, *(longlong *)local_a8 == 0)) {
    (**(code **)(*(longlong *)(local_a8 + 8) + 0x30))();
  }
  if ((-1 < *(int *)local_f8) &&
     (*(longlong *)local_f8 = *(longlong *)local_f8 + -1, *(longlong *)local_f8 == 0)) {
    (**(code **)(*(longlong *)(local_f8 + 8) + 0x30))();
  }
  if ((-1 < *(int *)local_e0) &&
     (*(longlong *)local_e0 = *(longlong *)local_e0 + -1, *(longlong *)local_e0 == 0)) {
    (**(code **)(*(longlong *)(local_e0 + 8) + 0x30))();
  }
  if ((-1 < *(int *)local_d8) &&
     (*(longlong *)local_d8 = *(longlong *)local_d8 + -1, *(longlong *)local_d8 == 0)) {
    (**(code **)(*(longlong *)(local_d8 + 8) + 0x30))();
  }
  if ((-1 < *(int *)local_a0) &&
     (*(longlong *)local_a0 = *(longlong *)local_a0 + -1, *(longlong *)local_a0 == 0)) {
    (**(code **)(*(longlong *)(local_a0 + 8) + 0x30))();
  }
  if ((-1 < (int)*local_98) && (*local_98 = *local_98 + -1, *local_98 == 0)) {
    (**(code **)(local_98[1] + 0x30))();
  }
  if ((-1 < *(int *)pcVar18) &&
     (*(longlong *)pcVar18 = *(longlong *)pcVar18 + -1, *(longlong *)pcVar18 == 0)) {
    (**(code **)(*(longlong *)(pcVar18 + 8) + 0x30))(pcVar18);
  }
  if ((-1 < *(int *)pcVar13) &&
     (*(longlong *)pcVar13 = *(longlong *)pcVar13 + -1, *(longlong *)pcVar13 == 0)) {
    (**(code **)(*(longlong *)(pcVar13 + 8) + 0x30))(pcVar13);
  }
  if ((-1 < *(int *)pcVar8) &&
     (*(longlong *)pcVar8 = *(longlong *)pcVar8 + -1, *(longlong *)pcVar8 == 0)) {
    (**(code **)(*(longlong *)(pcVar8 + 8) + 0x30))(pcVar8);
  }
  plVar10 = *(longlong **)(param_1 + 0x60);
  *(longlong **)(param_1 + 0x60) = plVar11;
  if (((plVar10 != (longlong *)0x0) && (-1 < (int)*plVar10)) &&
     (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
    (**(code **)(plVar10[1] + 0x30))();
  }
  return 0;
}
