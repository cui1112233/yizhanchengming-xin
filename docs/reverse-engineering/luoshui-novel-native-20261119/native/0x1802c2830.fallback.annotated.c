/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x1802c2830 */

longlong * FUN_1802c2830(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  int *piVar1;
  char cVar2;
  int iVar3;
  longlong *plVar4;
  longlong *plVar5;
  longlong lVar6;
  longlong *plVar7;
  code *pcVar8;
  longlong lVar9;
  longlong *plVar10;
  undefined8 uVar11;
  code *pcVar12;
  code *pcVar13;
  longlong *plVar14;
  longlong lVar15;
  longlong *plVar16;
  undefined4 uVar17;
  undefined4 local_res18;
  longlong *local_c8 [2];
  longlong *local_b8;
  code *local_b0;
  longlong *local_a8;
  longlong *local_a0;
  longlong *local_98;
  longlong *local_90;
  longlong *local_88;
  longlong *local_80;
  longlong *local_78;
  longlong *local_70;
  longlong *local_68;
  longlong *local_60;
  undefined8 local_58;
  longlong *local_50;
  code *local_48;
  longlong *local_40;
  
  plVar10 = DAT_1803c4b88;
  plVar7 = (longlong *)*param_3;
  plVar16 = (longlong *)0x0;
  local_78 = (longlong *)0x0;
  local_80 = (longlong *)0x0;
  local_b8 = (longlong *)0x0;
  local_b0 = (code *)0x0;
  local_a0 = (longlong *)0x0;
  local_a8 = (longlong *)0x0;
  local_c8[0] = DAT_180391e98;
  local_98 = DAT_1803c4b88;
  local_70 = plVar7;
  if (DAT_1803c4b88 == (longlong *)0x0) {
LAB_1802c28b8:
    DAT_1803c4b88 = (longlong *)FUN_18036cd10(param_1,DAT_1803c4bc0,DAT_1803ca468,0x38);
  }
  else if ((1 < *DAT_1803c4b88) || (DAT_1803c4b88[2] != 0)) {
    if ((DAT_1803c4b88 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803c4b88 && (*DAT_1803c4b88 = *DAT_1803c4b88 + -1, *plVar10 == 0)))) {
      (**(code **)(plVar10[1] + 0x30))(plVar10);
    }
    goto LAB_1802c28b8;
  }
  plVar10 = DAT_1803c4b88;
  local_98 = DAT_1803c4b88;
  FUN_180001d00(param_1,DAT_1803c4b88);
/* novels.services.spot_fix_patches#0="=== PATCHES ===" */
  iVar3 = PySequence_Contains(plVar7,DAT_1803c4c30);
  if (iVar3 == -1) {
    plVar5 = *(longlong **)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
    plVar16 = (longlong *)0x0;
    local_res18 = 0x18;
  }
  else if (iVar3 == 1) {
    plVar4 = (longlong *)FUN_18037f0c0();
    if (plVar4 == (longlong *)0x0) {
      plVar5 = *(longlong **)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      local_res18 = 0x18;
    }
    else {
      local_res18 = 0x18;
      *(undefined4 *)(plVar10 + 5) = 0x18;
      plVar5 = (longlong *)FUN_1803770d0(param_1,plVar4,DAT_1803c4c40);
      if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
        (**(code **)(plVar4[1] + 0x30))(plVar4);
      }
      if (plVar5 == (longlong *)0x0) {
        plVar5 = *(longlong **)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
      }
      else {
        plVar4 = (longlong *)FUN_180380b80(plVar5,DAT_1803c4c48);
        if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
          (**(code **)(plVar5[1] + 0x30))(plVar5);
        }
        if (plVar4 == (longlong *)0x0) {
          plVar5 = *(longlong **)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
        }
        else {
          plVar5 = (longlong *)FUN_18014b800(param_1,plVar4,_Py_NoneStruct_exref);
          if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
            (**(code **)(plVar4[1] + 0x30))(plVar4);
          }
          local_88 = (longlong *)FUN_180001b80(param_1,plVar7,plVar5);
          if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
            (**(code **)(plVar5[1] + 0x30))(plVar5);
          }
          if (local_88 != (longlong *)0x0) goto LAB_1802c2a60;
          plVar5 = *(longlong **)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
        }
      }
    }
  }
  else {
    iVar3 = (int)*plVar7 + 1;
    local_88 = plVar7;
    if (iVar3 != 0) {
      *(int *)plVar7 = iVar3;
    }
LAB_1802c2a60:
    local_78 = local_88;
    plVar16 = (longlong *)FUN_180372160(param_1,0);
    lVar6 = FUN_1802c23b0(param_1);
    if (lVar6 == 0) {
/* novels.services.spot_fix_patches#4="re" */
      FUN_180373340(param_1,local_c8,DAT_1803c4c50);
      plVar5 = local_c8[0];
      local_res18 = 0x1a;
    }
    else {
      plVar7 = (longlong *)FUN_18037f0c0();
/* novels.services.spot_fix_patches#6="--- PATCH(?:\\s+\\d+)? ---\\s*TARGET_TEXT:\\s*([\\s\\S]*?)\\s*REPLACEMENT_TEXT:\\s*([\\s\\S]*?)\\s*--- END PATCH ---" */
      uVar11 = DAT_1803c4c60;
      if (plVar7 == (longlong *)0x0) {
        plVar5 = *(longlong **)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
        local_res18 = 0x1a;
      }
      else {
        lVar6 = FUN_1802c23b0(param_1);
        if (lVar6 == 0) {
/* novels.services.spot_fix_patches#4="re" */
          FUN_180373340(param_1,local_c8,DAT_1803c4c50);
          if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
            (**(code **)(plVar7[1] + 0x30))(plVar7);
          }
          plVar5 = local_c8[0];
          local_res18 = 0x1c;
        }
        else {
          plVar4 = (longlong *)FUN_18037f0c0();
          if (plVar4 == (longlong *)0x0) {
            plVar5 = *(longlong **)(param_1 + 0x60);
            *(undefined8 *)(param_1 + 0x60) = 0;
            if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
              (**(code **)(plVar7[1] + 0x30))(plVar7);
            }
            local_res18 = 0x1c;
          }
          else {
            local_58 = uVar11;
            *(undefined4 *)(plVar10 + 5) = 0x1a;
            local_50 = plVar4;
            plVar5 = (longlong *)FUN_180377750(param_1,plVar7,&local_58);
            local_60 = plVar5;
            if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
              (**(code **)(plVar7[1] + 0x30))(plVar7);
            }
            if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
              (**(code **)(plVar4[1] + 0x30))(plVar4);
            }
            if (plVar5 == (longlong *)0x0) {
              plVar5 = *(longlong **)(param_1 + 0x60);
              *(undefined8 *)(param_1 + 0x60) = 0;
              local_res18 = 0x1a;
            }
            else {
              *(undefined4 *)(plVar10 + 5) = 0x1e;
              local_res18 = 0x1e;
              local_80 = plVar5;
/* novels.services.spot_fix_patches#8="finditer" */
              plVar7 = (longlong *)FUN_18037e6c0(param_1,plVar5,DAT_1803c4c70,local_88);
              if (plVar7 == (longlong *)0x0) {
                plVar5 = *(longlong **)(param_1 + 0x60);
                *(undefined8 *)(param_1 + 0x60) = 0;
                local_res18 = 0x1e;
              }
              else {
                local_90 = (longlong *)FUN_180001930();
                if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                  (**(code **)(plVar7[1] + 0x30))(plVar7);
                }
                plVar7 = local_a0;
                if (local_90 == (longlong *)0x0) {
                  plVar5 = *(longlong **)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                  local_res18 = 0x1e;
                }
                else {
                  do {
                    plVar4 = local_a8;
                    pcVar13 = local_b0;
                    plVar10 = local_b8;
                    pcVar12 = *(code **)(local_90[1] + 0xe0);
                    local_68 = plVar7;
                    if (pcVar12 == (code *)0x0) {
                      PyErr_Format(*(undefined8 *)PyExc_TypeError_exref,
                                   "\'%s\' object is not an iterator",
                                   *(undefined8 *)(local_90[1] + 0x18));
LAB_1802c3344:
                      plVar5 = *(longlong **)(param_1 + 0x60);
                      if (plVar5 != (longlong *)0x0) {
                        plVar14 = (longlong *)plVar5[1];
                        if ((*(uint *)((longlong *)plVar5[1] + 0x15) & 0x40000000) == 0) {
                          plVar14 = plVar5;
                        }
                        if ((plVar14 != *(longlong **)PyExc_StopIteration_exref) &&
                           (((-1 < *(int *)(plVar14[1] + 0xa8) ||
                             ((*(uint *)(plVar14 + 0x15) & 0x40000000) == 0)) ||
                            (cVar2 = FUN_180382640(), cVar2 != '\x01')))) {
                          *(undefined8 *)(param_1 + 0x60) = 0;
                          uVar17 = 0x1e;
                          local_b8 = plVar10;
                          local_b0 = pcVar13;
                          local_a8 = plVar4;
                          local_a0 = plVar7;
                          if (plVar4 != (longlong *)0x0) goto LAB_1802c33b9;
                          goto LAB_1802c33d5;
                        }
                        *(undefined8 *)(param_1 + 0x60) = 0;
                        if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                          (**(code **)(plVar5[1] + 0x30))(plVar5);
                        }
                      }
                      if (((plVar4 != (longlong *)0x0) && (-1 < (int)*plVar4)) &&
                         (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
                        (**(code **)(plVar4[1] + 0x30))(plVar4);
                      }
                      if ((-1 < (int)*local_90) && (*local_90 = *local_90 + -1, *local_90 == 0)) {
                        (**(code **)(local_90[1] + 0x30))();
                      }
                      if (plVar16 != (longlong *)0x0) {
                        iVar3 = (int)*plVar16 + 1;
                        if (iVar3 != 0) {
                          *(int *)plVar16 = iVar3;
                        }
                        FUN_180001d80(param_1);
                        if ((-1 < (int)*plVar16) && (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                          (**(code **)(plVar16[1] + 0x30))(plVar16);
                        }
                        if ((-1 < (int)*local_88) && (*local_88 = *local_88 + -1, *local_88 == 0)) {
                          (**(code **)(local_88[1] + 0x30))();
                        }
                        if ((-1 < (int)*local_60) && (*local_60 = *local_60 + -1, *local_60 == 0)) {
                          (**(code **)(local_80[1] + 0x30))();
                        }
                        if (((plVar10 != (longlong *)0x0) && (-1 < (int)*plVar10)) &&
                           (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                          (**(code **)(plVar10[1] + 0x30))(plVar10);
                        }
                        if (((pcVar13 != (code *)0x0) && (-1 < *(int *)pcVar13)) &&
                           (*(longlong *)pcVar13 = *(longlong *)pcVar13 + -1,
                           *(longlong *)pcVar13 == 0)) {
                          (**(code **)(*(longlong *)(pcVar13 + 8) + 0x30))(pcVar13);
                        }
                        if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) &&
                           (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                          (**(code **)(plVar7[1] + 0x30))(plVar7);
                        }
                        if ((int)*local_70 < 0) {
                          return plVar16;
                        }
                        *local_70 = *local_70 + -1;
                        if (*local_70 != 0) {
                          return plVar16;
                        }
                        (**(code **)(local_70[1] + 0x30))();
                        return plVar16;
                      }
                      plVar7 = (longlong *)
                               PyUnicode_FromFormat
                                         ("cannot access local variable \'%U\' where it is not associated with a value"
/* novels.services.spot_fix_patches#14="patches" */
                                          ,DAT_1803c4ca0);
                      plVar5 = (longlong *)
                               FUN_180376770(param_1,*(undefined8 *)PyExc_UnboundLocalError_exref,
                                             plVar7);
                      if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                        (**(code **)(plVar7[1] + 0x30))(plVar7);
                      }
                      FUN_18000d1d0(param_1,plVar5);
                      local_res18 = 0x23;
                      plVar10 = local_98;
                      goto LAB_1802c34b6;
                    }
                    plVar5 = (longlong *)(*pcVar12)();
                    plVar7 = local_a0;
                    if (plVar5 == (longlong *)0x0) goto LAB_1802c3344;
                    local_a8 = plVar5;
                    if (((plVar4 != (longlong *)0x0) && (-1 < (int)*plVar4)) &&
                       (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
                      (**(code **)(plVar4[1] + 0x30))(plVar4);
                    }
                    iVar3 = (int)*plVar5 + 1;
                    if (iVar3 != 0) {
                      *(int *)plVar5 = iVar3;
                    }
                    local_b8 = plVar5;
                    if (((plVar10 != (longlong *)0x0) && (-1 < (int)*plVar10)) &&
                       (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                      (**(code **)(plVar10[1] + 0x30))();
                    }
                    lVar6 = FUN_1802c2250(param_1);
                    plVar7 = local_98;
                    if (lVar6 == 0) {
/* novels.services.spot_fix_patches#9="_trim_field" */
                      FUN_180373340(param_1,local_c8,DAT_1803c4c78);
                      uVar17 = 0x1f;
                      plVar5 = local_c8[0];
                      goto LAB_1802c33b9;
                    }
                    *(undefined4 *)(local_98 + 5) = 0x1f;
/* novels.services.spot_fix_patches#10="group" */
                    pcVar12 = (code *)FUN_18037e6c0(param_1,plVar5,DAT_1803c4c80,
                                                    *(undefined8 *)(DAT_1803c4c88 + 0x18));
                    if (pcVar12 == (code *)0x0) {
                      plVar5 = *(longlong **)(param_1 + 0x60);
                      uVar17 = 0x1f;
                      *(undefined8 *)(param_1 + 0x60) = 0;
                      goto LAB_1802c33b9;
                    }
                    iVar3 = FUN_180001510(pcVar12);
                    if (iVar3 == -1) {
                      plVar5 = *(longlong **)(param_1 + 0x60);
                      *(undefined8 *)(param_1 + 0x60) = 0;
                      if ((-1 < *(int *)pcVar12) &&
                         (*(longlong *)pcVar12 = *(longlong *)pcVar12 + -1,
                         *(longlong *)pcVar12 == 0)) {
                        (**(code **)(*(longlong *)(pcVar12 + 8) + 0x30))(pcVar12);
                      }
                      uVar17 = 0x1f;
                      goto LAB_1802c33b9;
                    }
                    if (iVar3 != 1) {
                      if ((-1 < *(int *)pcVar12) &&
                         (*(longlong *)pcVar12 = *(longlong *)pcVar12 + -1,
                         *(longlong *)pcVar12 == 0)) {
                        (**(code **)(*(longlong *)(pcVar12 + 8) + 0x30))(pcVar12);
                      }
/* novels.services.spot_fix_patches#12="" */
                      pcVar12 = DAT_1803c4c90;
/* novels.services.spot_fix_patches#12="" */
                      if (*(int *)DAT_1803c4c90 + 1 != 0) {
/* novels.services.spot_fix_patches#12="" */
                        *(int *)DAT_1803c4c90 = *(int *)DAT_1803c4c90 + 1;
                      }
                    }
                    *(undefined4 *)(plVar7 + 5) = 0x1f;
                    pcVar8 = (code *)FUN_180376770(param_1,lVar6,pcVar12);
                    if ((-1 < *(int *)pcVar12) &&
                       (*(longlong *)pcVar12 = *(longlong *)pcVar12 + -1, *(longlong *)pcVar12 == 0)
                       ) {
                      (**(code **)(*(longlong *)(pcVar12 + 8) + 0x30))(pcVar12);
                    }
                    if (pcVar8 == (code *)0x0) {
                      plVar5 = *(longlong **)(param_1 + 0x60);
                      uVar17 = 0x1f;
                      *(undefined8 *)(param_1 + 0x60) = 0;
                      goto LAB_1802c33b9;
                    }
                    local_b0 = pcVar8;
                    if (((pcVar13 != (code *)0x0) && (-1 < *(int *)pcVar13)) &&
                       (*(longlong *)pcVar13 = *(longlong *)pcVar13 + -1, *(longlong *)pcVar13 == 0)
                       ) {
                      (**(code **)(*(longlong *)(pcVar13 + 8) + 0x30))(pcVar13);
                    }
                    lVar6 = FUN_1802c2250(param_1);
                    if (lVar6 == 0) {
/* novels.services.spot_fix_patches#9="_trim_field" */
                      FUN_180373340(param_1,local_c8,DAT_1803c4c78);
                      uVar17 = 0x20;
                      plVar5 = local_c8[0];
                      goto LAB_1802c33b9;
                    }
                    *(undefined4 *)(plVar7 + 5) = 0x20;
/* novels.services.spot_fix_patches#10="group" */
                    pcVar12 = (code *)FUN_18037e6c0(param_1,plVar5,DAT_1803c4c80,
                                                    *(undefined8 *)(DAT_1803c4c98 + 0x18));
                    if (pcVar12 == (code *)0x0) {
LAB_1802c3265:
                      plVar5 = *(longlong **)(param_1 + 0x60);
                      uVar17 = 0x20;
                      *(undefined8 *)(param_1 + 0x60) = 0;
                      goto LAB_1802c33b9;
                    }
                    if (pcVar12 != _Py_TrueStruct_exref) {
                      if ((pcVar12 != _Py_FalseStruct_exref) && (pcVar12 != _Py_NoneStruct_exref)) {
                        lVar9 = *(longlong *)(pcVar12 + 8);
                        if ((*(longlong *)(lVar9 + 0x60) == 0) ||
                           (pcVar13 = *(code **)(*(longlong *)(lVar9 + 0x60) + 0x48),
                           pcVar13 == (code *)0x0)) {
                          if (((*(longlong **)(lVar9 + 0x70) == (longlong *)0x0) ||
                              (pcVar13 = (code *)**(longlong **)(lVar9 + 0x70),
                              pcVar13 == (code *)0x0)) &&
                             ((*(longlong **)(lVar9 + 0x68) == (longlong *)0x0 ||
                              (pcVar13 = (code *)**(longlong **)(lVar9 + 0x68),
                              pcVar13 == (code *)0x0)))) goto LAB_1802c2e93;
                          lVar9 = (*pcVar13)(pcVar12);
                        }
                        else {
                          iVar3 = (*pcVar13)(pcVar12);
                          lVar9 = (longlong)iVar3;
                        }
                        if (0 < lVar9) goto LAB_1802c2e93;
                        if (lVar9 != 0) {
                          plVar5 = *(longlong **)(param_1 + 0x60);
                          *(undefined8 *)(param_1 + 0x60) = 0;
                          if ((-1 < *(int *)pcVar12) &&
                             (*(longlong *)pcVar12 = *(longlong *)pcVar12 + -1,
                             *(longlong *)pcVar12 == 0)) {
                            (**(code **)(*(longlong *)(pcVar12 + 8) + 0x30))(pcVar12);
                          }
                          uVar17 = 0x20;
                          goto LAB_1802c33b9;
                        }
                      }
                      if ((-1 < *(int *)pcVar12) &&
                         (*(longlong *)pcVar12 = *(longlong *)pcVar12 + -1,
                         *(longlong *)pcVar12 == 0)) {
                        (**(code **)(*(longlong *)(pcVar12 + 8) + 0x30))(pcVar12);
                      }
/* novels.services.spot_fix_patches#12="" */
                      pcVar12 = DAT_1803c4c90;
/* novels.services.spot_fix_patches#12="" */
                      if (*(int *)DAT_1803c4c90 + 1 != 0) {
/* novels.services.spot_fix_patches#12="" */
                        *(int *)DAT_1803c4c90 = *(int *)DAT_1803c4c90 + 1;
                      }
                    }
LAB_1802c2e93:
                    *(undefined4 *)(plVar7 + 5) = 0x20;
                    plVar7 = (longlong *)FUN_180376770(param_1,lVar6,pcVar12);
                    if ((-1 < *(int *)pcVar12) &&
                       (*(longlong *)pcVar12 = *(longlong *)pcVar12 + -1, *(longlong *)pcVar12 == 0)
                       ) {
                      (**(code **)(*(longlong *)(pcVar12 + 8) + 0x30))(pcVar12);
                    }
                    if (plVar7 == (longlong *)0x0) goto LAB_1802c3265;
                    plVar10 = plVar7;
                    if (((local_a0 != (longlong *)0x0) && (-1 < (int)*local_68)) &&
                       (*local_68 = *local_68 + -1, *local_68 == 0)) {
                      plVar10 = local_a0 + 1;
                      local_a0 = plVar7;
                      (**(code **)(*plVar10 + 0x30))();
                      plVar10 = local_a0;
                    }
                    local_a0 = plVar10;
                    if (pcVar8 == _Py_TrueStruct_exref) {
LAB_1802c2f90:
                      if (plVar16 == (longlong *)0x0) {
                        plVar7 = (longlong *)
                                 PyUnicode_FromFormat
                                           ("cannot access local variable \'%U\' where it is not associated with a value"
/* novels.services.spot_fix_patches#14="patches" */
                                            ,DAT_1803c4ca0);
                        plVar5 = (longlong *)
                                 FUN_180376770(param_1,*(undefined8 *)PyExc_UnboundLocalError_exref,
                                               plVar7);
                        if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                          (**(code **)(plVar7[1] + 0x30))(plVar7);
                        }
                        FUN_18000d1d0(param_1,plVar5);
                        uVar17 = 0x22;
                        goto LAB_1802c33b9;
                      }
                      lVar6 = *(longlong *)(DAT_1803ca470 + 0x20);
                      if (*(char *)(lVar6 + 10) == '\0') {
                        plVar10 = (longlong *)FUN_1800015a0();
                        if (plVar10 == (longlong *)0x0) {
                          lVar9 = 0;
                        }
                        else {
                          lVar9 = *plVar10;
                        }
LAB_1802c3063:
                        if (lVar9 == 0) goto LAB_1802c3068;
                      }
                      else {
                        iVar3 = *(int *)(lVar6 + 0xc);
                        if (iVar3 == 0) {
                          iVar3 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
                          *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar3 + 1;
                          *(int *)(lVar6 + 0xc) = iVar3;
                        }
                        if (iVar3 != DAT_1803a9ed8) {
                          DAT_1803a9ed8 = iVar3;
                          DAT_1803c4b60 =
/* novels.services.spot_fix_patches#15="SpotFixPatch" */
                               FUN_1803707d0(lVar6,DAT_1803c4ca8,
/* novels.services.spot_fix_patches#15="SpotFixPatch" */
                                             *(undefined8 *)(DAT_1803c4ca8 + 0x18));
                        }
                        if (-1 < DAT_1803c4b60) {
                          lVar15 = (1L << (*(byte *)(lVar6 + 9) & 0x3f)) + lVar6;
                          lVar9 = *(longlong *)(lVar15 + 0x28 + DAT_1803c4b60 * 0x10);
                          if (lVar9 != 0) goto LAB_1802c3090;
                          DAT_1803c4b60 =
/* novels.services.spot_fix_patches#15="SpotFixPatch" */
                               FUN_1803707d0(lVar6,DAT_1803c4ca8,
/* novels.services.spot_fix_patches#15="SpotFixPatch" */
                                             *(undefined8 *)(DAT_1803c4ca8 + 0x18));
                          if (-1 < DAT_1803c4b60) {
                            lVar9 = *(longlong *)(lVar15 + 0x28 + DAT_1803c4b60 * 0x10);
                            goto LAB_1802c3063;
                          }
                        }
LAB_1802c3068:
/* novels.services.spot_fix_patches#15="SpotFixPatch" */
                        plVar10 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803c4ca8);
                        if ((plVar10 == (longlong *)0x0) || (lVar9 = *plVar10, lVar9 == 0)) {
/* novels.services.spot_fix_patches#15="SpotFixPatch" */
                          FUN_180373340(param_1,local_c8,DAT_1803c4ca8);
                          uVar17 = 0x22;
                          plVar5 = local_c8[0];
                          goto LAB_1802c33b9;
                        }
                      }
LAB_1802c3090:
                      *(undefined4 *)(local_98 + 5) = 0x22;
                      local_48 = pcVar8;
                      local_40 = plVar7;
                      lVar6 = FUN_18037ccb0(param_1,lVar9,&local_48,DAT_1803c4cb0);
                      if (lVar6 == 0) {
                        plVar5 = *(longlong **)(param_1 + 0x60);
                        uVar17 = 0x22;
                        *(undefined8 *)(param_1 + 0x60) = 0;
                        goto LAB_1802c33b9;
                      }
                      lVar9 = plVar16[2];
                      cVar2 = FUN_1803722b0(plVar16,lVar9 + 1);
                      if (cVar2 != '\0') {
                        *(longlong *)(plVar16[3] + lVar9 * 8) = lVar6;
                      }
                    }
                    else if ((pcVar8 != _Py_FalseStruct_exref) && (pcVar8 != _Py_NoneStruct_exref))
                    {
                      lVar6 = *(longlong *)(pcVar8 + 8);
                      if ((*(longlong *)(lVar6 + 0x60) == 0) ||
                         (pcVar12 = *(code **)(*(longlong *)(lVar6 + 0x60) + 0x48),
                         pcVar12 == (code *)0x0)) {
                        if ((*(undefined8 **)(lVar6 + 0x70) == (undefined8 *)0x0) ||
                           (pcVar12 = (code *)**(undefined8 **)(lVar6 + 0x70),
                           pcVar12 == (code *)0x0)) {
                          if ((*(undefined8 **)(lVar6 + 0x68) == (undefined8 *)0x0) ||
                             (pcVar12 = (code *)**(undefined8 **)(lVar6 + 0x68),
                             pcVar12 == (code *)0x0)) goto LAB_1802c2f90;
                          lVar6 = (*pcVar12)(pcVar8);
                        }
                        else {
                          lVar6 = (*pcVar12)(pcVar8);
                        }
                      }
                      else {
                        iVar3 = (*pcVar12)(pcVar8);
                        lVar6 = (longlong)iVar3;
                      }
                      if (0 < lVar6) goto LAB_1802c2f90;
                      if (lVar6 != 0) {
                        plVar5 = *(longlong **)(param_1 + 0x60);
                        *(undefined8 *)(param_1 + 0x60) = 0;
                        uVar17 = 0x21;
                        goto LAB_1802c33b9;
                      }
                    }
                    lVar6 = *(longlong *)(param_1 + 0x10);
                    if ((((*(int *)(*(longlong *)(lVar6 + 0x60) + 0x4d4) != 0) ||
                         (*(int *)(lVar6 + 0x1a0) != 0)) &&
                        (iVar3 = Py_MakePendingCalls(), iVar3 < 0)) &&
                       (*(longlong *)(param_1 + 0x60) != 0)) goto LAB_1802c31d8;
                    if (*(int *)(lVar6 + 0x174) != 0) {
                      PyEval_SaveThread();
                      PyEval_AcquireThread(param_1);
                    }
                  } while (*(longlong *)(param_1 + 0x80) == 0);
                  plVar7 = *(longlong **)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x80) = 0;
                  uVar11 = FUN_180373490(param_1);
                  *(undefined8 *)(param_1 + 0x60) = uVar11;
                  if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) &&
                     (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                    (**(code **)(plVar7[1] + 0x30))(plVar7);
                  }
LAB_1802c31d8:
                  plVar5 = *(longlong **)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                  uVar17 = local_res18;
LAB_1802c33b9:
                  local_res18 = uVar17;
                  if ((-1 < (int)*local_a8) && (*local_a8 = *local_a8 + -1, *local_a8 == 0)) {
                    (**(code **)(local_a8[1] + 0x30))(local_a8);
                  }
LAB_1802c33d5:
                  plVar10 = local_98;
                  if ((-1 < (int)*local_90) && (*local_90 = *local_90 + -1, *local_90 == 0)) {
                    (**(code **)(local_90[1] + 0x30))();
                    plVar10 = local_98;
                  }
                }
              }
            }
          }
        }
      }
    }
  }
LAB_1802c34b6:
  piVar1 = (int *)plVar5[5];
  if (piVar1 == (int *)0x0) {
    pcVar12 = (code *)FUN_18038e2a0(plVar10,local_res18);
    plVar7 = (longlong *)plVar5[5];
    pcVar13 = pcVar12;
    if (pcVar12 == _Py_NoneStruct_exref) {
      pcVar13 = (code *)0x0;
    }
    else {
LAB_1802c3510:
      pcVar12 = pcVar13;
      if ((pcVar13 != (code *)0x0) && (*(int *)pcVar13 + 1 != 0)) {
        *(int *)pcVar13 = *(int *)pcVar13 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar1 + 6) == plVar10) goto LAB_1802c3559;
    pcVar12 = (code *)FUN_18038e2a0(plVar10,local_res18);
    *(int **)(pcVar12 + 0x10) = piVar1;
    if (*piVar1 + 1 != 0) {
      *piVar1 = *piVar1 + 1;
    }
    plVar7 = (longlong *)plVar5[5];
    pcVar13 = pcVar12;
    if (pcVar12 != _Py_NoneStruct_exref) goto LAB_1802c3510;
    pcVar13 = (code *)0x0;
  }
  plVar5[5] = (longlong)pcVar13;
  if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)
     ) {
    (**(code **)(plVar7[1] + 0x30))();
  }
  if (((pcVar12 != (code *)0x0) && (-1 < *(int *)pcVar12)) &&
     (*(longlong *)pcVar12 = *(longlong *)pcVar12 + -1, *(longlong *)pcVar12 == 0)) {
    (**(code **)(*(longlong *)(pcVar12 + 8) + 0x30))(pcVar12);
  }
LAB_1802c3559:
  plVar14 = local_a0;
  pcVar12 = local_b0;
  plVar4 = local_b8;
  FUN_18036d0a0(plVar10,"ooooooo",local_70,plVar16,local_78,local_80,local_b8,local_b0,local_a0);
  plVar7 = DAT_1803c4b88;
  if (plVar10 == DAT_1803c4b88) {
    if ((-1 < (int)*DAT_1803c4b88) && (*DAT_1803c4b88 = *DAT_1803c4b88 + -1, *plVar7 == 0)) {
      (**(code **)(plVar7[1] + 0x30))();
    }
    DAT_1803c4b88 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if (((plVar16 != (longlong *)0x0) && (-1 < (int)*plVar16)) &&
     (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
    (**(code **)(plVar16[1] + 0x30))(plVar16);
  }
  if (((local_78 != (longlong *)0x0) && (-1 < (int)*local_78)) &&
     (*local_78 = *local_78 + -1, *local_78 == 0)) {
    (**(code **)(local_78[1] + 0x30))();
  }
  if (((local_80 != (longlong *)0x0) && (-1 < (int)*local_80)) &&
     (*local_80 = *local_80 + -1, *local_80 == 0)) {
    (**(code **)(local_80[1] + 0x30))();
  }
  if (((plVar4 != (longlong *)0x0) && (-1 < (int)*plVar4)) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)
     ) {
    (**(code **)(plVar4[1] + 0x30))(plVar4);
  }
  if (((pcVar12 != (code *)0x0) && (-1 < *(int *)pcVar12)) &&
     (*(longlong *)pcVar12 = *(longlong *)pcVar12 + -1, *(longlong *)pcVar12 == 0)) {
    (**(code **)(*(longlong *)(pcVar12 + 8) + 0x30))(pcVar12);
  }
  if (((plVar14 != (longlong *)0x0) && (-1 < (int)*plVar14)) &&
     (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
    (**(code **)(plVar14[1] + 0x30))(plVar14);
  }
  if ((-1 < (int)*local_70) && (*local_70 = *local_70 + -1, *local_70 == 0)) {
    (**(code **)(local_70[1] + 0x30))();
  }
  plVar7 = *(longlong **)(param_1 + 0x60);
  *(longlong **)(param_1 + 0x60) = plVar5;
  if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)
     ) {
    (**(code **)(plVar7[1] + 0x30))();
  }
  return (longlong *)0x0;
}
