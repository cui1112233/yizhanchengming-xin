/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x1800d22e0 */

longlong * FUN_1800d22e0(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  int *piVar2;
  undefined8 *puVar3;
  undefined8 uVar4;
  int iVar5;
  longlong lVar6;
  longlong *plVar7;
  longlong *plVar8;
  longlong *plVar9;
  code *pcVar10;
  code *pcVar11;
  longlong *plVar12;
  longlong *plVar13;
  longlong *plVar14;
  undefined8 uVar15;
  longlong lVar16;
  longlong lVar17;
  longlong *local_res18;
  longlong local_a8;
  longlong *local_a0;
  longlong *local_98;
  longlong *local_90;
  longlong *local_88;
  longlong *local_80;
  longlong *local_78;
  longlong *local_70;
  longlong *local_68;
  longlong *local_60;
  longlong *local_58;
  longlong *local_50;
  longlong *local_48;
  longlong *local_40;

  plVar14 = DAT_1803b2ab8;
  local_90 = (longlong *)*param_3;
  plVar1 = (longlong *)param_3[6];
  local_88 = (longlong *)param_3[1];
  local_80 = (longlong *)param_3[2];
  local_78 = (longlong *)param_3[3];
  local_70 = (longlong *)param_3[4];
  local_68 = (longlong *)param_3[5];
  local_a8 = DAT_180391d00;
  local_98 = (longlong *)0x0;
  local_res18 = (longlong *)0x0;
  local_a0 = DAT_1803b2ab8;
  if (DAT_1803b2ab8 == (longlong *)0x0) {
LAB_1800d2384:
    DAT_1803b2ab8 = (longlong *)FUN_18036cd10(param_1,DAT_1803b2950,DAT_1803ca7a0,0x48);
  }
  else if ((1 < *DAT_1803b2ab8) || (DAT_1803b2ab8[2] != 0)) {
    if ((DAT_1803b2ab8 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b2ab8 && (*DAT_1803b2ab8 = *DAT_1803b2ab8 + -1, *plVar14 == 0)))) {
      (**(code **)(plVar14[1] + 0x30))(plVar14);
    }
    goto LAB_1800d2384;
  }
  plVar14 = DAT_1803b2ab8;
  local_a0 = DAT_1803b2ab8;
  FUN_180001d00(param_1,DAT_1803b2ab8);
  lVar6 = FUN_1800c7020(param_1);
  if (lVar6 == 0) {
/* novels.pipeline.imitation_short_runner#4="FileRepository" */
    FUN_180373340(param_1,&local_a8,DAT_1803b2b20);
    uVar15 = 0x185;
    lVar6 = local_a8;
    local_res18 = (longlong *)0x0;
  }
  else {
    plVar7 = (longlong *)FUN_18037f0c0();
    if (plVar7 == (longlong *)0x0) {
      lVar6 = *(longlong *)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      uVar15 = 0x185;
      local_res18 = (longlong *)0x0;
    }
    else {
/* novels.pipeline.imitation_short_runner#50="imitation" */
      plVar8 = (longlong *)FUN_180382540(plVar1,DAT_1803b2c90);
      if (plVar8 == (longlong *)0x0) {
        lVar6 = *(longlong *)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
        if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
          (**(code **)(plVar7[1] + 0x30))(plVar7);
        }
        uVar15 = 0x185;
        local_res18 = (longlong *)0x0;
      }
      else {
/* novels.pipeline.imitation_short_runner#142="blueprint.md" */
        plVar9 = (longlong *)FUN_180382540(plVar8,DAT_1803b2f70);
        if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
          (**(code **)(plVar8[1] + 0x30))(plVar8);
        }
        if (plVar9 == (longlong *)0x0) {
          lVar6 = *(longlong *)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
            (**(code **)(plVar7[1] + 0x30))(plVar7);
          }
          uVar15 = 0x185;
        }
        else {
          uVar15 = 0x185;
/* novels.pipeline.imitation_short_runner#2="" */
          local_58 = DAT_1803b2b10;
          *(undefined4 *)(plVar14 + 5) = 0x185;
          local_60 = plVar9;
          plVar8 = (longlong *)FUN_180377750(param_1,plVar7,&local_60);
          if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
            (**(code **)(plVar7[1] + 0x30))(plVar7);
          }
          if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
            (**(code **)(plVar9[1] + 0x30))(plVar9);
          }
          if (plVar8 != (longlong *)0x0) {
            plVar7 = (longlong *)FUN_18037f0c0();
            if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
              (**(code **)(plVar8[1] + 0x30))(plVar8);
            }
            if (plVar7 != (longlong *)0x0) {
              *(undefined4 *)(plVar14 + 5) = 0x185;
              plVar8 = (longlong *)FUN_180376210(param_1,plVar7);
              if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                (**(code **)(plVar7[1] + 0x30))(plVar7);
              }
              if (plVar8 != (longlong *)0x0) {
                local_98 = plVar8;
                iVar5 = FUN_180001510(plVar8);
                if (iVar5 == -1) {
                  uVar15 = 0x186;
                }
                else if (iVar5 == 0) {
                  uVar15 = 0x18b;
                  *(undefined4 *)(plVar14 + 5) = 0x18b;
                  plVar7 = (longlong *)
/* novels.pipeline.imitation_short_runner#124="_stage" ; novels.pipeline.imitation_short_runner#170={"type": "tuple", "items": [{"tag": "l", "fileRange": [3955506, 3955508], "value": 2}, {"tag": "a", "fileRange": [3955508, 3955516], "value": "active"}, {"tag": "u", "fileRange": [3955516, 3955536], "value": "构建创作蓝图"}]} */
                           FUN_18037e980(param_1,local_90,DAT_1803b2ee0,DAT_1803b3050 + 0x18);
                  if (plVar7 != (longlong *)0x0) {
                    if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                      (**(code **)(plVar7[1] + 0x30))(plVar7);
                    }
                    plVar7 = (longlong *)FUN_18037f0c0();
/* novels.pipeline.imitation_short_runner#151="architect" */
                    piVar2 = DAT_1803b2fb8;
                    if (plVar7 != (longlong *)0x0) {
                      lVar6 = *(longlong *)(param_1 + 0x10);
                      plVar9 = *(longlong **)(lVar6 + 0x41da0);
                      if (plVar9 == (longlong *)0x0) {
                        plVar9 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,2);
                      }
                      else {
                        *(longlong *)(lVar6 + 0x41da0) = plVar9[3];
                        *(int *)(lVar6 + 0x41e3c) = *(int *)(lVar6 + 0x41e3c) + -1;
                        *plVar9 = 1;
                      }
                      plVar12 = plVar9 + -2;
                      plVar9[3] = 0;
                      plVar9[4] = 0;
                      lVar6 = _PyThreadState_GetCurrent();
                      lVar6 = *(longlong *)(*(longlong *)(lVar6 + 0x10) + 0xd0);
                      puVar3 = *(undefined8 **)(lVar6 + 8);
                      *puVar3 = plVar12;
                      plVar9[-1] = plVar9[-1] & 3;
                      plVar9[-1] = plVar9[-1] | (ulonglong)puVar3;
                      *plVar12 = lVar6;
                      *(longlong **)(lVar6 + 8) = plVar12;
                      if (*piVar2 + 1 != 0) {
                        *piVar2 = *piVar2 + 1;
                      }
                      plVar9[3] = (longlong)piVar2;
/* novels.pipeline.imitation_short_runner#153="system" */
                      uVar4 = DAT_1803b2fc8;
/* novels.pipeline.imitation_short_runner#152="role" */
                      uVar15 = DAT_1803b2fc0;
                      plVar12 = (longlong *)_PyDict_NewPresized(2);
                      PyDict_SetItem(plVar12,uVar15,uVar4);
/* novels.pipeline.imitation_short_runner#154="content" */
                      uVar15 = DAT_1803b2fd0;
                      lVar6 = *(longlong *)(DAT_1803ca798 + 0x20);
                      if (*(char *)(lVar6 + 10) == '\0') {
                        plVar13 = (longlong *)FUN_1800015a0();
                        if (plVar13 == (longlong *)0x0) {
                          lVar6 = 0;
                        }
                        else {
                          lVar6 = *plVar13;
                        }
LAB_1800d28b3:
                        if (lVar6 == 0) goto LAB_1800d28b8;
LAB_1800d28f6:
                        *(undefined4 *)(plVar14 + 5) = 399;
                        plVar13 = (longlong *)FUN_180376210(param_1);
                        if (plVar13 == (longlong *)0x0) {
                          lVar6 = *(longlong *)(param_1 + 0x60);
                          *(undefined8 *)(param_1 + 0x60) = 0;
                          goto LAB_1800d2917;
                        }
                        PyDict_SetItem(plVar12,uVar15,plVar13);
                        if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                          (**(code **)(plVar13[1] + 0x30))(plVar13);
                        }
                        plVar13 = (longlong *)FUN_180372160(param_1,2);
                        *(longlong **)plVar13[3] = plVar12;
/* novels.pipeline.imitation_short_runner#156="user" */
                        uVar4 = DAT_1803b2fe0;
/* novels.pipeline.imitation_short_runner#152="role" */
                        uVar15 = DAT_1803b2fc0;
                        plVar14 = (longlong *)_PyDict_NewPresized(2);
                        PyDict_SetItem(plVar14,uVar15,uVar4);
/* novels.pipeline.imitation_short_runner#154="content" */
                        uVar15 = DAT_1803b2fd0;
                        lVar6 = *(longlong *)(DAT_1803ca798 + 0x20);
                        if (*(char *)(lVar6 + 10) == '\0') {
                          plVar12 = (longlong *)FUN_1800015a0();
                          if (plVar12 == (longlong *)0x0) {
                            lVar16 = 0;
                          }
                          else {
                            lVar16 = *plVar12;
                          }
LAB_1800d2a7b:
                          if (lVar16 == 0) goto LAB_1800d2a80;
LAB_1800d2abe:
                          *(undefined4 *)(local_a0 + 5) = 400;
                          local_60 = local_88;
                          local_58 = local_80;
                          local_50 = local_78;
                          local_48 = local_70;
                          local_40 = local_68;
                          plVar12 = (longlong *)
/* novels.pipeline.imitation_short_runner#173={"type": "tuple", "items": [{"tag": "a", "fileRange": [3955598, 3955617], "value": "target_word_count"}, {"tag": "a", "fileRange": [3955617, 3955637], "value": "reader_preferences"}, {"tag": "a", "fileRange": [3955637, 3955657], "value": "supplementary_info"}]} */
                                    FUN_18037d650(param_1,lVar16,&local_60,&local_50,DAT_1803b3068);
                          if (plVar12 != (longlong *)0x0) {
                            PyDict_SetItem(plVar14,uVar15,plVar12);
                            if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                              (**(code **)(plVar12[1] + 0x30))(plVar12);
                            }
                            *(longlong **)(plVar13[3] + 8) = plVar14;
                            plVar9[4] = (longlong)plVar13;
/* novels.pipeline.imitation_short_runner#174={"type": "dict", "keys": [{"tag": "a", "fileRange": [3955659, 3955672], "value": "stage_label"}, {"tag": "a", "fileRange": [3955672, 3955684], "value": "max_tokens"}, {"tag": "a", "fileRange": [3955684, 3955697], "value": "temperature"}], "values": [{"tag": "u", "fileRange": [3955697, 3955717], "value": "构建创作蓝图"}, {"tag": "l", "fileRange": [3955717, 3955720], "value": 8192}, {"tag": "f", "fileRange": [3955720, 3955729], "value": 0.5}]} */
                            plVar12 = (longlong *)FUN_18036fb80(param_1,DAT_1803b3070);
                            plVar14 = local_a0;
                            uVar15 = 0x18c;
                            *(undefined4 *)(local_a0 + 5) = 0x18c;
                            plVar13 = (longlong *)FUN_180001790(param_1,plVar7,plVar9,plVar12);
                            local_a0 = plVar13;
                            if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                              (**(code **)(plVar7[1] + 0x30))(plVar7);
                            }
                            if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                              (**(code **)(plVar9[1] + 0x30))(plVar9);
                            }
                            if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                              (**(code **)(plVar12[1] + 0x30))(plVar12);
                            }
                            if (plVar13 != (longlong *)0x0) {
                              lVar6 = FUN_1800c7020(param_1);
                              local_res18 = plVar13;
                              if (lVar6 == 0) {
/* novels.pipeline.imitation_short_runner#4="FileRepository" */
                                FUN_180373340(param_1,&local_a8,DAT_1803b2b20);
                                uVar15 = 0x19c;
                                lVar6 = local_a8;
                              }
                              else {
                                plVar7 = (longlong *)FUN_18037f0c0();
                                if (plVar7 == (longlong *)0x0) {
                                  lVar6 = *(longlong *)(param_1 + 0x60);
                                  uVar15 = 0x19c;
                                  *(undefined8 *)(param_1 + 0x60) = 0;
                                }
                                else {
/* novels.pipeline.imitation_short_runner#50="imitation" */
                                  plVar9 = (longlong *)FUN_180382540(plVar1,DAT_1803b2c90);
                                  if (plVar9 == (longlong *)0x0) {
                                    lVar6 = *(longlong *)(param_1 + 0x60);
                                    *(undefined8 *)(param_1 + 0x60) = 0;
                                    if ((-1 < (int)*plVar7) &&
                                       (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                      (**(code **)(plVar7[1] + 0x30))(plVar7);
                                    }
                                    uVar15 = 0x19c;
                                  }
                                  else {
/* novels.pipeline.imitation_short_runner#142="blueprint.md" */
                                    plVar12 = (longlong *)FUN_180382540(plVar9,DAT_1803b2f70);
                                    if ((-1 < (int)*plVar9) &&
                                       (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                      (**(code **)(plVar9[1] + 0x30))(plVar9);
                                    }
                                    local_res18 = local_a0;
                                    if (plVar12 == (longlong *)0x0) {
                                      lVar6 = *(longlong *)(param_1 + 0x60);
                                      *(undefined8 *)(param_1 + 0x60) = 0;
                                      if ((-1 < (int)*plVar7) &&
                                         (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                        (**(code **)(plVar7[1] + 0x30))(plVar7);
                                      }
                                      uVar15 = 0x19c;
                                      local_res18 = plVar13;
                                    }
                                    else {
                                      uVar15 = 0x19c;
                                      local_58 = local_a0;
                                      *(undefined4 *)(plVar14 + 5) = 0x19c;
                                      local_60 = plVar12;
                                      plVar9 = (longlong *)FUN_180377750(param_1,plVar7,&local_60);
                                      if ((-1 < (int)*plVar7) &&
                                         (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                        (**(code **)(plVar7[1] + 0x30))(plVar7);
                                      }
                                      if ((-1 < (int)*plVar12) &&
                                         (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                        (**(code **)(plVar12[1] + 0x30))(plVar12);
                                      }
                                      if (plVar9 != (longlong *)0x0) {
                                        if ((-1 < (int)*plVar9) &&
                                           (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                          (**(code **)(plVar9[1] + 0x30))(plVar9);
                                        }
                                        plVar7 = local_90;
                                        uVar15 = 0x19d;
                                        *(undefined4 *)(plVar14 + 5) = 0x19d;
                                        plVar9 = (longlong *)
/* novels.pipeline.imitation_short_runner#124="_stage" */
                                                 FUN_18037e820(param_1,local_90,DAT_1803b2ee0,
/* novels.pipeline.imitation_short_runner#175={"type": "tuple", "items": [{"tag": "l", "fileRange": [3955731, 3955733], "value": 2}, {"tag": "a", "fileRange": [3955733, 3955744], "value": "completed"}]} */
                                                               DAT_1803b3078 + 0x18);
                                        if (plVar9 != (longlong *)0x0) {
                                          if ((-1 < (int)*plVar9) &&
                                             (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                            (**(code **)(plVar9[1] + 0x30))(plVar9);
                                          }
                                          FUN_180001d80(param_1);
                                          iVar5 = (int)*local_res18 + 1;
                                          plVar14 = local_res18;
                                          if (iVar5 != 0) {
                                            *(int *)local_res18 = iVar5;
                                          }
LAB_1800d2e51:
                                          if ((-1 < (int)*plVar8) &&
                                             (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                            (**(code **)(plVar8[1] + 0x30))(plVar8);
                                          }
                                          if (((plVar14 != (longlong *)0x0) && (-1 < (int)*plVar14))
                                             && (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                                            (**(code **)(plVar14[1] + 0x30))(plVar14);
                                          }
                                          if ((-1 < (int)*plVar7) &&
                                             (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                            (**(code **)(plVar7[1] + 0x30))(plVar7);
                                          }
                                          if ((-1 < (int)*local_88) &&
                                             (*local_88 = *local_88 + -1, *local_88 == 0)) {
                                            (**(code **)(local_88[1] + 0x30))();
                                          }
                                          if ((-1 < (int)*local_80) &&
                                             (*local_80 = *local_80 + -1, *local_80 == 0)) {
                                            (**(code **)(local_80[1] + 0x30))();
                                          }
                                          if ((-1 < (int)*local_78) &&
                                             (*local_78 = *local_78 + -1, *local_78 == 0)) {
                                            (**(code **)(local_78[1] + 0x30))();
                                          }
                                          if ((-1 < (int)*local_70) &&
                                             (*local_70 = *local_70 + -1, *local_70 == 0)) {
                                            (**(code **)(local_70[1] + 0x30))();
                                          }
                                          if ((-1 < (int)*local_68) &&
                                             (*local_68 = *local_68 + -1, *local_68 == 0)) {
                                            (**(code **)(local_68[1] + 0x30))();
                                          }
                                          if ((int)*plVar1 < 0) {
                                            return local_res18;
                                          }
                                          *plVar1 = *plVar1 + -1;
                                          if (*plVar1 != 0) {
                                            return local_res18;
                                          }
                                          (**(code **)(plVar1[1] + 0x30))(plVar1);
                                          return local_res18;
                                        }
                                      }
                                      lVar6 = *(longlong *)(param_1 + 0x60);
                                      *(undefined8 *)(param_1 + 0x60) = 0;
                                    }
                                  }
                                }
                              }
                              goto LAB_1800d25ad;
                            }
                            goto LAB_1800d2599;
                          }
                          lVar6 = *(longlong *)(param_1 + 0x60);
                          *(undefined8 *)(param_1 + 0x60) = 0;
                        }
                        else {
                          iVar5 = *(int *)(lVar6 + 0xc);
                          if (iVar5 == 0) {
                            iVar5 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
                            *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar5 + 1;
                            *(int *)(lVar6 + 0xc) = iVar5;
                          }
                          if (iVar5 != DAT_1803a9788) {
                            DAT_1803a9788 = iVar5;
                            DAT_1803b2a10 =
/* novels.pipeline.imitation_short_runner#172="build_blueprint_user_prompt" */
                                 FUN_1803707d0(lVar6,DAT_1803b3060,
/* novels.pipeline.imitation_short_runner#172="build_blueprint_user_prompt" */
                                               *(undefined8 *)(DAT_1803b3060 + 0x18));
                          }
                          if (-1 < DAT_1803b2a10) {
                            lVar17 = (1L << (*(byte *)(lVar6 + 9) & 0x3f)) + lVar6;
                            lVar16 = *(longlong *)(lVar17 + 0x28 + DAT_1803b2a10 * 0x10);
                            if (lVar16 != 0) goto LAB_1800d2abe;
                            DAT_1803b2a10 =
/* novels.pipeline.imitation_short_runner#172="build_blueprint_user_prompt" */
                                 FUN_1803707d0(lVar6,DAT_1803b3060,
/* novels.pipeline.imitation_short_runner#172="build_blueprint_user_prompt" */
                                               *(undefined8 *)(DAT_1803b3060 + 0x18));
                            if (-1 < DAT_1803b2a10) {
                              lVar16 = *(longlong *)(lVar17 + 0x28 + DAT_1803b2a10 * 0x10);
                              goto LAB_1800d2a7b;
                            }
                          }
LAB_1800d2a80:
/* novels.pipeline.imitation_short_runner#172="build_blueprint_user_prompt" */
                          plVar12 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803b3060);
                          if ((plVar12 != (longlong *)0x0) && (lVar16 = *plVar12, lVar16 != 0))
                          goto LAB_1800d2abe;
/* novels.pipeline.imitation_short_runner#172="build_blueprint_user_prompt" */
                          FUN_180373340(param_1,&local_a8,DAT_1803b3060);
                          lVar6 = local_a8;
                        }
                        uVar15 = 400;
                        if ((-1 < (int)*plVar14) && (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                          (**(code **)(plVar14[1] + 0x30))(plVar14);
                        }
                        plVar14 = local_a0;
                        if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                          (**(code **)(plVar13[1] + 0x30))(plVar13);
                          plVar14 = local_a0;
                        }
                      }
                      else {
                        iVar5 = *(int *)(lVar6 + 0xc);
                        if (iVar5 == 0) {
                          iVar5 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
                          *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar5 + 1;
                          *(int *)(lVar6 + 0xc) = iVar5;
                        }
                        if (iVar5 != DAT_1803a9780) {
                          DAT_1803a9780 = iVar5;
                          DAT_1803b2a68 =
/* novels.pipeline.imitation_short_runner#171="build_blueprint_system_prompt" */
                               FUN_1803707d0(lVar6,DAT_1803b3058,
/* novels.pipeline.imitation_short_runner#171="build_blueprint_system_prompt" */
                                             *(undefined8 *)(DAT_1803b3058 + 0x18));
                        }
                        if (-1 < DAT_1803b2a68) {
                          lVar16 = (1L << (*(byte *)(lVar6 + 9) & 0x3f)) + lVar6;
                          if (*(longlong *)(lVar16 + 0x28 + DAT_1803b2a68 * 0x10) != 0)
                          goto LAB_1800d28f6;
                          DAT_1803b2a68 =
/* novels.pipeline.imitation_short_runner#171="build_blueprint_system_prompt" */
                               FUN_1803707d0(lVar6,DAT_1803b3058,
/* novels.pipeline.imitation_short_runner#171="build_blueprint_system_prompt" */
                                             *(undefined8 *)(DAT_1803b3058 + 0x18));
                          if (-1 < DAT_1803b2a68) {
                            lVar6 = *(longlong *)(lVar16 + 0x28 + DAT_1803b2a68 * 0x10);
                            goto LAB_1800d28b3;
                          }
                        }
LAB_1800d28b8:
/* novels.pipeline.imitation_short_runner#171="build_blueprint_system_prompt" */
                        plVar13 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803b3058);
                        if ((plVar13 != (longlong *)0x0) && (*plVar13 != 0)) goto LAB_1800d28f6;
/* novels.pipeline.imitation_short_runner#171="build_blueprint_system_prompt" */
                        FUN_180373340(param_1,&local_a8,DAT_1803b3058);
                        lVar6 = local_a8;
LAB_1800d2917:
                        uVar15 = 399;
                        if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                          (**(code **)(plVar12[1] + 0x30))(plVar12);
                        }
                      }
                      if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                        (**(code **)(plVar7[1] + 0x30))(plVar7);
                      }
                      if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                        (**(code **)(plVar9[1] + 0x30))(plVar9);
                      }
                      goto LAB_1800d25ad;
                    }
                    uVar15 = 0x18c;
                  }
                }
                else {
                  lVar6 = FUN_1800c7610();
                  if (lVar6 == 0) {
/* novels.pipeline.imitation_short_runner#145="logger" */
                    FUN_180373340(param_1,&local_a8,DAT_1803b2f88);
                    uVar15 = 0x187;
                    lVar6 = local_a8;
                    goto LAB_1800d25ad;
                  }
                  uVar15 = 0x187;
                  *(undefined4 *)(plVar14 + 5) = 0x187;
                  plVar7 = (longlong *)
/* novels.pipeline.imitation_short_runner#146="info" */
                           FUN_18037e6c0(param_1,lVar6,DAT_1803b2f90,
/* novels.pipeline.imitation_short_runner#168={"type": "tuple", "items": [{"tag": "u", "fileRange": [3955436, 3955472], "value": "Resuming: blueprint already exists"}]} */
                                         *(undefined8 *)(DAT_1803b3040 + 0x18));
                  if (plVar7 != (longlong *)0x0) {
                    if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                      (**(code **)(plVar7[1] + 0x30))(plVar7);
                    }
                    plVar7 = local_90;
                    uVar15 = 0x188;
                    *(undefined4 *)(plVar14 + 5) = 0x188;
                    plVar9 = (longlong *)
/* novels.pipeline.imitation_short_runner#124="_stage" ; novels.pipeline.imitation_short_runner#169={"type": "tuple", "items": [{"tag": "l", "fileRange": [3955474, 3955476], "value": 2}, {"tag": "a", "fileRange": [3955476, 3955487], "value": "completed"}, {"tag": "u", "fileRange": [3955487, 3955504], "value": "（已缓存）"}]} */
                             FUN_18037e980(param_1,local_90,DAT_1803b2ee0,DAT_1803b3048 + 0x18);
                    if (plVar9 != (longlong *)0x0) {
                      if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                        (**(code **)(plVar9[1] + 0x30))(plVar9);
                      }
                      iVar5 = (int)*plVar8 + 1;
                      if (iVar5 != 0) {
                        *(int *)plVar8 = iVar5;
                      }
                      FUN_180001d80(param_1);
                      local_res18 = plVar8;
                      plVar14 = (longlong *)0x0;
                      goto LAB_1800d2e51;
                    }
                  }
                }
              }
            }
          }
LAB_1800d2599:
          lVar6 = *(longlong *)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
        }
      }
    }
  }
LAB_1800d25ad:
  piVar2 = *(int **)(lVar6 + 0x28);
  if (piVar2 == (int *)0x0) {
    pcVar10 = (code *)FUN_18038e2a0(plVar14,uVar15);
    plVar7 = *(longlong **)(lVar6 + 0x28);
    pcVar11 = pcVar10;
    if (pcVar10 == _Py_NoneStruct_exref) {
      pcVar11 = (code *)0x0;
    }
    else {
LAB_1800d2f91:
      pcVar10 = pcVar11;
      if ((pcVar11 != (code *)0x0) && (*(int *)pcVar11 + 1 != 0)) {
        *(int *)pcVar11 = *(int *)pcVar11 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar2 + 6) == plVar14) goto LAB_1800d2fda;
    pcVar10 = (code *)FUN_18038e2a0(plVar14,uVar15);
    *(int **)(pcVar10 + 0x10) = piVar2;
    if (*piVar2 + 1 != 0) {
      *piVar2 = *piVar2 + 1;
    }
    plVar7 = *(longlong **)(lVar6 + 0x28);
    pcVar11 = pcVar10;
    if (pcVar10 != _Py_NoneStruct_exref) goto LAB_1800d2f91;
    pcVar11 = (code *)0x0;
  }
  *(code **)(lVar6 + 0x28) = pcVar11;
  if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)
     ) {
    (**(code **)(plVar7[1] + 0x30))();
  }
  if (((pcVar10 != (code *)0x0) && (-1 < *(int *)pcVar10)) &&
     (*(longlong *)pcVar10 = *(longlong *)pcVar10 + -1, *(longlong *)pcVar10 == 0)) {
    (**(code **)(*(longlong *)(pcVar10 + 8) + 0x30))(pcVar10);
  }
LAB_1800d2fda:
  plVar8 = local_90;
  FUN_18036d0a0(plVar14,"ooooooooo",local_90,local_88,local_80,local_78,local_70,local_68,plVar1,
                local_98,local_res18);
  plVar7 = DAT_1803b2ab8;
  if (plVar14 == DAT_1803b2ab8) {
    if ((-1 < (int)*DAT_1803b2ab8) && (*DAT_1803b2ab8 = *DAT_1803b2ab8 + -1, *plVar7 == 0)) {
      (**(code **)(plVar7[1] + 0x30))();
    }
    DAT_1803b2ab8 = (longlong *)0x0;
  }
  lVar16 = **(longlong **)(param_1 + 0x38);
  plVar14 = *(longlong **)(lVar16 + 0x30);
  **(longlong **)(param_1 + 0x38) = *(longlong *)(lVar16 + 8);
  *(undefined4 *)(plVar14 + 8) = 0xffffffff;
  plVar7 = (longlong *)plVar14[2];
  if (((plVar7 != (longlong *)0x0) && (plVar14[2] = 0, -1 < (int)*plVar7)) &&
     (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
    (**(code **)(plVar7[1] + 0x30))();
  }
  if ((-1 < (int)*plVar14) && (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
    (**(code **)(plVar14[1] + 0x30))(plVar14);
  }
  plVar14[10] = 0;
  if (((local_98 != (longlong *)0x0) && (-1 < (int)*local_98)) &&
     (*local_98 = *local_98 + -1, *local_98 == 0)) {
    (**(code **)(local_98[1] + 0x30))();
  }
  if (((local_res18 != (longlong *)0x0) && (-1 < (int)*local_res18)) &&
     (*local_res18 = *local_res18 + -1, *local_res18 == 0)) {
    (**(code **)(local_res18[1] + 0x30))(local_res18);
  }
  if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
    (**(code **)(plVar8[1] + 0x30))(plVar8);
  }
  if ((-1 < (int)*local_88) && (*local_88 = *local_88 + -1, *local_88 == 0)) {
    (**(code **)(local_88[1] + 0x30))();
  }
  if ((-1 < (int)*local_80) && (*local_80 = *local_80 + -1, *local_80 == 0)) {
    (**(code **)(local_80[1] + 0x30))();
  }
  if ((-1 < (int)*local_78) && (*local_78 = *local_78 + -1, *local_78 == 0)) {
    (**(code **)(local_78[1] + 0x30))();
  }
  if ((-1 < (int)*local_70) && (*local_70 = *local_70 + -1, *local_70 == 0)) {
    (**(code **)(local_70[1] + 0x30))();
  }
  if ((-1 < (int)*local_68) && (*local_68 = *local_68 + -1, *local_68 == 0)) {
    (**(code **)(local_68[1] + 0x30))();
  }
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))(plVar1);
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong *)(param_1 + 0x60) = lVar6;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return (longlong *)0x0;
}
