/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x18014cdf0 */

longlong * FUN_18014cdf0(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  undefined8 *puVar1;
  int *piVar2;
  longlong *plVar3;
  char cVar4;
  int iVar5;
  longlong lVar6;
  longlong *plVar7;
  longlong *plVar8;
  longlong *plVar9;
  longlong *plVar10;
  undefined8 uVar11;
  code *pcVar12;
  code *pcVar13;
  uint uVar14;
  longlong *plVar15;
  longlong lVar16;
  undefined8 uVar17;
  longlong *plVar18;
  longlong *local_res18;
  longlong *local_e0 [2];
  longlong *local_d0;
  longlong *local_c8;
  longlong *local_c0;
  longlong *local_b8;
  longlong *local_b0;
  longlong *local_a8;
  longlong *local_a0;
  longlong *local_90;
  longlong *local_88;
  longlong *local_80;
  longlong *local_78;
  longlong *local_70;
  longlong *local_68;
  longlong *local_60;
  longlong *local_58;
  undefined8 local_50;
  longlong *local_48;

  plVar3 = DAT_1803b6b68;
  local_a0 = (longlong *)*param_3;
  plVar7 = (longlong *)param_3[1];
  plVar18 = (longlong *)0x0;
  local_e0[0] = DAT_180391d20;
  local_b0 = (longlong *)0x0;
  local_90 = (longlong *)0x0;
  local_88 = (longlong *)0x0;
  local_a8 = (longlong *)0x0;
  local_d0 = (longlong *)0x0;
  local_c8 = (longlong *)0x0;
  local_res18 = (longlong *)0x0;
  local_c0 = (longlong *)0x0;
  local_b8 = (longlong *)0x0;
  local_70 = plVar7;
  if (DAT_1803b6b68 == (longlong *)0x0) {
LAB_18014ce87:
    DAT_1803b6b68 = (longlong *)FUN_18036cd10(param_1,DAT_1803b6a50,DAT_1803ca758,0x50);
  }
  else if ((1 < *DAT_1803b6b68) || (DAT_1803b6b68[2] != 0)) {
    if ((DAT_1803b6b68 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b6b68 && (*DAT_1803b6b68 = *DAT_1803b6b68 + -1, *plVar3 == 0)))) {
      (**(code **)(plVar3[1] + 0x30))(plVar3);
    }
    goto LAB_18014ce87;
  }
  plVar3 = DAT_1803b6b68;
  FUN_180001d00(param_1,DAT_1803b6b68);
  lVar6 = FUN_18014c500(param_1);
  if (lVar6 == 0) {
/* novels.pipeline.short_fiction_parsers#0="re" */
    FUN_180373340(param_1,local_e0,DAT_1803b6c30);
    uVar17 = 0x3c;
    plVar7 = local_e0[0];
    local_res18 = (longlong *)0x0;
  }
  else {
    uVar17 = 0x3c;
    *(undefined4 *)(plVar3 + 5) = 0x3c;
/* novels.pipeline.short_fiction_parsers#1="escape" */
    plVar7 = (longlong *)FUN_18037e6c0(param_1,lVar6,DAT_1803b6c38,plVar7);
    local_68 = plVar7;
    if (plVar7 == (longlong *)0x0) {
      plVar7 = *(longlong **)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      local_res18 = plVar18;
    }
    else {
      local_90 = plVar7;
      lVar6 = FUN_18014c500(param_1);
      if (lVar6 == 0) {
/* novels.pipeline.short_fiction_parsers#0="re" */
        FUN_180373340(param_1,local_e0,DAT_1803b6c30);
        uVar17 = 0x3d;
        plVar7 = local_e0[0];
        local_res18 = plVar18;
      }
      else {
        plVar8 = (longlong *)FUN_18037f0c0();
/* novels.pipeline.short_fiction_parsers#3="^\\s*===\\s*" */
        piVar2 = DAT_1803b6c48;
        if (plVar8 == (longlong *)0x0) {
          plVar7 = *(longlong **)(param_1 + 0x60);
          uVar17 = 0x3d;
          *(undefined8 *)(param_1 + 0x60) = 0;
          local_res18 = plVar18;
        }
        else {
          lVar6 = *(longlong *)(param_1 + 0x10);
          plVar10 = *(longlong **)(lVar6 + 0x41da8);
          if (plVar10 == (longlong *)0x0) {
            plVar10 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,3);
          }
          else {
            *(longlong *)(lVar6 + 0x41da8) = plVar10[3];
            *(int *)(lVar6 + 0x41e40) = *(int *)(lVar6 + 0x41e40) + -1;
            *plVar10 = 1;
          }
          plVar9 = plVar10 + -2;
          plVar10[3] = 0;
          plVar10[4] = 0;
          plVar10[5] = 0;
          lVar6 = _PyThreadState_GetCurrent();
          lVar6 = *(longlong *)(*(longlong *)(lVar6 + 0x10) + 0xd0);
          puVar1 = *(undefined8 **)(lVar6 + 8);
          *puVar1 = plVar9;
          plVar10[-1] = plVar10[-1] & 3;
          plVar10[-1] = plVar10[-1] | (ulonglong)puVar1;
          *plVar9 = lVar6;
          *(longlong **)(lVar6 + 8) = plVar9;
          if (*piVar2 + 1 != 0) {
            *piVar2 = *piVar2 + 1;
          }
          plVar10[3] = (longlong)piVar2;
/* novels.pipeline.short_fiction_parsers#4="" */
          lVar6 = FUN_180375520(param_1,plVar7,DAT_1803b6c50);
          if (lVar6 == 0) {
            plVar7 = *(longlong **)(param_1 + 0x60);
            uVar17 = 0x3d;
            *(undefined8 *)(param_1 + 0x60) = 0;
            if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
              (**(code **)(plVar8[1] + 0x30))(plVar8);
            }
            local_res18 = (longlong *)0x0;
            if ((-1 < (int)*plVar10) &&
               (*plVar10 = *plVar10 + -1, local_res18 = plVar18, *plVar10 == 0)) {
              (**(code **)(plVar10[1] + 0x30))(plVar10);
            }
          }
          else {
            plVar10[4] = lVar6;
/* novels.pipeline.short_fiction_parsers#5="\\s*===\\s*$" */
            piVar2 = DAT_1803b6c58;
/* novels.pipeline.short_fiction_parsers#5="\\s*===\\s*$" */
            if (*DAT_1803b6c58 + 1 != 0) {
/* novels.pipeline.short_fiction_parsers#5="\\s*===\\s*$" */
              *DAT_1803b6c58 = *DAT_1803b6c58 + 1;
            }
            plVar10[5] = (longlong)piVar2;
/* novels.pipeline.short_fiction_parsers#4="" */
            plVar9 = (longlong *)PyUnicode_Join(DAT_1803b6c50,plVar10);
            if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
              (**(code **)(plVar10[1] + 0x30))(plVar10);
            }
            if (plVar9 == (longlong *)0x0) {
              plVar7 = *(longlong **)(param_1 + 0x60);
              *(undefined8 *)(param_1 + 0x60) = 0;
              if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                (**(code **)(plVar8[1] + 0x30))(plVar8);
              }
LAB_18014d0e7:
              uVar17 = 0x3d;
              local_res18 = plVar18;
            }
            else {
              lVar6 = FUN_18014c500(param_1);
              if (lVar6 == 0) {
/* novels.pipeline.short_fiction_parsers#0="re" */
                FUN_180373340(param_1,local_e0,DAT_1803b6c30);
                if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                  (**(code **)(plVar8[1] + 0x30))(plVar8);
                }
                if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                  (**(code **)(plVar9[1] + 0x30))(plVar9);
                }
LAB_18014d143:
                uVar17 = 0x3d;
                plVar7 = local_e0[0];
                local_res18 = plVar18;
              }
              else {
                plVar10 = (longlong *)FUN_18037f0c0();
                if (plVar10 == (longlong *)0x0) {
                  plVar7 = *(longlong **)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                  if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                    (**(code **)(plVar8[1] + 0x30))(plVar8);
                  }
                  if (((int)*plVar9 < 0) || (*plVar9 = *plVar9 + -1, *plVar9 != 0))
                  goto LAB_18014d0e7;
                  (**(code **)(plVar9[1] + 0x30))(plVar9);
                  uVar17 = 0x3d;
                  local_res18 = plVar18;
                }
                else {
                  lVar6 = FUN_18014c500(param_1);
                  if (lVar6 == 0) {
/* novels.pipeline.short_fiction_parsers#0="re" */
                    FUN_180373340(param_1,local_e0,DAT_1803b6c30);
                    if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                      (**(code **)(plVar8[1] + 0x30))(plVar8);
                    }
                    if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                      (**(code **)(plVar9[1] + 0x30))(plVar9);
                    }
                    if (((int)*plVar10 < 0) || (*plVar10 = *plVar10 + -1, *plVar10 != 0))
                    goto LAB_18014d143;
                    (**(code **)(plVar10[1] + 0x30))(plVar10);
                    uVar17 = 0x3d;
                    plVar7 = local_e0[0];
                    local_res18 = plVar18;
                  }
                  else {
                    plVar7 = (longlong *)FUN_18037f0c0();
                    if (plVar7 == (longlong *)0x0) {
                      plVar7 = *(longlong **)(param_1 + 0x60);
                      *(undefined8 *)(param_1 + 0x60) = 0;
                      if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                        (**(code **)(plVar8[1] + 0x30))(plVar8);
                      }
                      if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                        (**(code **)(plVar9[1] + 0x30))(plVar9);
                      }
                      if (((int)*plVar10 < 0) || (*plVar10 = *plVar10 + -1, *plVar10 != 0))
                      goto LAB_18014d0e7;
                      (**(code **)(plVar10[1] + 0x30))(plVar10);
                      uVar17 = 0x3d;
                      local_res18 = plVar18;
                    }
                    else {
                      plVar18 = (longlong *)FUN_180381230(plVar10,plVar7);
                      if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                        (**(code **)(plVar10[1] + 0x30))(plVar10);
                      }
                      if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                        (**(code **)(plVar7[1] + 0x30))(plVar7);
                      }
                      if (plVar18 == (longlong *)0x0) {
                        plVar7 = *(longlong **)(param_1 + 0x60);
                        *(undefined8 *)(param_1 + 0x60) = 0;
                        if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                          (**(code **)(plVar8[1] + 0x30))(plVar8);
                        }
                        if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                          (**(code **)(plVar9[1] + 0x30))(plVar9);
                        }
                        uVar17 = 0x3d;
                      }
                      else {
                        uVar17 = 0x3d;
                        *(undefined4 *)(plVar3 + 5) = 0x3d;
                        local_60 = plVar9;
                        local_58 = plVar18;
                        plVar7 = (longlong *)FUN_180377750(param_1,plVar8,&local_60);
                        local_80 = plVar7;
                        if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                          (**(code **)(plVar8[1] + 0x30))(plVar8);
                        }
                        if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                          (**(code **)(plVar9[1] + 0x30))(plVar9);
                        }
                        if ((-1 < (int)*plVar18) && (*plVar18 = *plVar18 + -1, *plVar18 == 0)) {
                          (**(code **)(plVar18[1] + 0x30))(plVar18);
                        }
                        if (plVar7 == (longlong *)0x0) {
LAB_18014d3b3:
                          plVar7 = *(longlong **)(param_1 + 0x60);
                          *(undefined8 *)(param_1 + 0x60) = 0;
                        }
                        else {
                          local_88 = plVar7;
                          lVar6 = FUN_18014c500(param_1);
                          if (lVar6 == 0) {
/* novels.pipeline.short_fiction_parsers#0="re" */
                            FUN_180373340(param_1,local_e0,DAT_1803b6c30);
                            uVar17 = 0x3e;
                            plVar7 = local_e0[0];
                            local_res18 = (longlong *)0x0;
                          }
                          else {
                            plVar18 = (longlong *)FUN_18037f0c0();
/* novels.pipeline.short_fiction_parsers#8="^\\s*===\\s*[A-Z0-9_ ]+\\s*===\\s*$" */
                            uVar11 = DAT_1803b6c70;
                            if (plVar18 == (longlong *)0x0) {
                              plVar7 = *(longlong **)(param_1 + 0x60);
                              uVar17 = 0x3e;
                              *(undefined8 *)(param_1 + 0x60) = 0;
                              local_res18 = (longlong *)0x0;
                            }
                            else {
                              lVar6 = FUN_18014c500(param_1);
                              if (lVar6 == 0) {
/* novels.pipeline.short_fiction_parsers#0="re" */
                                FUN_180373340(param_1,local_e0,DAT_1803b6c30);
                                plVar7 = local_e0[0];
                                if ((-1 < (int)*plVar18) &&
                                   (*plVar18 = *plVar18 + -1, *plVar18 == 0)) {
                                  lVar6 = plVar18[1];
                                  plVar8 = plVar18;
LAB_18014d469:
                                  (**(code **)(lVar6 + 0x30))(plVar8);
                                  plVar7 = local_e0[0];
                                }
                              }
                              else {
                                plVar8 = (longlong *)FUN_18037f0c0();
                                if (plVar8 == (longlong *)0x0) {
                                  plVar7 = *(longlong **)(param_1 + 0x60);
                                  *(undefined8 *)(param_1 + 0x60) = 0;
                                  if ((-1 < (int)*plVar18) &&
                                     (*plVar18 = *plVar18 + -1, *plVar18 == 0)) {
                                    (**(code **)(plVar18[1] + 0x30))(plVar18);
                                    uVar17 = 0x3e;
                                    local_res18 = (longlong *)0x0;
                                    goto LAB_18014dd8e;
                                  }
                                }
                                else {
                                  lVar6 = FUN_18014c500(param_1);
                                  if (lVar6 != 0) {
                                    plVar7 = (longlong *)FUN_18037f0c0();
                                    if (plVar7 == (longlong *)0x0) {
                                      plVar7 = *(longlong **)(param_1 + 0x60);
                                      *(undefined8 *)(param_1 + 0x60) = 0;
                                      if ((-1 < (int)*plVar18) &&
                                         (*plVar18 = *plVar18 + -1, *plVar18 == 0)) {
                                        (**(code **)(plVar18[1] + 0x30))(plVar18);
                                      }
                                      if (((int)*plVar8 < 0) ||
                                         (*plVar8 = *plVar8 + -1, *plVar8 != 0)) goto LAB_18014d474;
                                      (**(code **)(plVar8[1] + 0x30))(plVar8);
                                      uVar17 = 0x3e;
                                      local_res18 = (longlong *)0x0;
                                    }
                                    else {
                                      plVar10 = (longlong *)FUN_180381230(plVar8,plVar7);
                                      if ((-1 < (int)*plVar8) &&
                                         (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                        (**(code **)(plVar8[1] + 0x30))(plVar8);
                                      }
                                      if ((-1 < (int)*plVar7) &&
                                         (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                        (**(code **)(plVar7[1] + 0x30))(plVar7);
                                      }
                                      if (plVar10 == (longlong *)0x0) {
                                        plVar7 = *(longlong **)(param_1 + 0x60);
                                        *(undefined8 *)(param_1 + 0x60) = 0;
                                        if ((-1 < (int)*plVar18) &&
                                           (*plVar18 = *plVar18 + -1, *plVar18 == 0)) {
                                          (**(code **)(plVar18[1] + 0x30))(plVar18);
                                        }
                                        uVar17 = 0x3e;
                                      }
                                      else {
                                        uVar17 = 0x3e;
                                        local_50 = uVar11;
                                        *(undefined4 *)(plVar3 + 5) = 0x3e;
                                        local_48 = plVar10;
                                        plVar7 = (longlong *)
                                                 FUN_180377750(param_1,plVar18,&local_50);
                                        local_78 = plVar7;
                                        if ((-1 < (int)*plVar18) &&
                                           (*plVar18 = *plVar18 + -1, *plVar18 == 0)) {
                                          (**(code **)(plVar18[1] + 0x30))(plVar18);
                                        }
                                        if ((-1 < (int)*plVar10) &&
                                           (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                          (**(code **)(plVar10[1] + 0x30))(plVar10);
                                        }
                                        if (plVar7 == (longlong *)0x0) {
                                          plVar7 = *(longlong **)(param_1 + 0x60);
                                          *(undefined8 *)(param_1 + 0x60) = 0;
                                        }
                                        else {
                                          local_a8 = plVar7;
                                          local_b0 = (longlong *)FUN_180372160(param_1,0);
                                          uVar17 = 0x40;
                                          *(undefined4 *)(plVar3 + 5) = 0x40;
                                          plVar7 = (longlong *)
/* novels.pipeline.short_fiction_parsers#9="finditer" */
                                                   FUN_18037e6c0(param_1,local_80,DAT_1803b6c78,
                                                                 local_a0);
                                          if (plVar7 == (longlong *)0x0) goto LAB_18014d3b3;
                                          plVar18 = (longlong *)FUN_18000d280();
                                          if ((-1 < (int)*plVar7) &&
                                             (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                            (**(code **)(plVar7[1] + 0x30))(plVar7);
                                          }
                                          if (plVar18 == (longlong *)0x0) {
                                            plVar7 = *(longlong **)(param_1 + 0x60);
                                            *(undefined8 *)(param_1 + 0x60) = 0;
                                            local_res18 = (longlong *)0x0;
                                          }
                                          else {
                                            do {
                                              plVar9 = local_b8;
                                              plVar10 = local_c8;
                                              plVar8 = local_d0;
                                              pcVar12 = *(code **)(plVar18[1] + 0xe0);
                                              local_e0[0] = local_res18;
                                              if (pcVar12 == (code *)0x0) {
                                                PyErr_Format(*(undefined8 *)PyExc_TypeError_exref,
                                                             "\'%s\' object is not an iterator",
                                                             *(undefined8 *)(plVar18[1] + 0x18));
LAB_18014dc0f:
                                                plVar7 = *(longlong **)(param_1 + 0x60);
                                                if (plVar7 != (longlong *)0x0) {
                                                  plVar15 = (longlong *)plVar7[1];
                                                  if ((*(uint *)((longlong *)plVar7[1] + 0x15) &
                                                      0x40000000) == 0) {
                                                    plVar15 = plVar7;
                                                  }
                                                  if ((plVar15 !=
                                                       *(longlong **)PyExc_StopIteration_exref) &&
                                                     (((-1 < *(int *)(plVar15[1] + 0xa8) ||
                                                       ((*(uint *)(plVar15 + 0x15) & 0x40000000) ==
                                                        0)) || (cVar4 = FUN_180382640(),
                                                               cVar4 != '\x01')))) {
                                                    *(undefined8 *)(param_1 + 0x60) = 0;
                                                    local_d0 = plVar8;
                                                    local_c8 = plVar10;
                                                    local_b8 = plVar9;
                                                    if (plVar9 != (longlong *)0x0)
                                                    goto LAB_18014dc89;
                                                    goto LAB_18014dca5;
                                                  }
                                                  *(undefined8 *)(param_1 + 0x60) = 0;
                                                  if ((-1 < (int)*plVar7) &&
                                                     (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                    (**(code **)(plVar7[1] + 0x30))(plVar7);
                                                  }
                                                }
                                                if (((plVar9 != (longlong *)0x0) &&
                                                    (-1 < (int)*plVar9)) &&
                                                   (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                  (**(code **)(plVar9[1] + 0x30))(plVar9);
                                                }
                                                if ((-1 < (int)*plVar18) &&
                                                   (*plVar18 = *plVar18 + -1, *plVar18 == 0)) {
                                                  (**(code **)(plVar18[1] + 0x30))(plVar18);
                                                }
                                                plVar7 = local_b0;
                                                if (local_b0 != (longlong *)0x0) {
                                                  iVar5 = (int)*local_b0 + 1;
                                                  if (iVar5 != 0) {
                                                    *(int *)local_b0 = iVar5;
                                                  }
                                                  FUN_180001d80(param_1);
                                                  if ((-1 < (int)*plVar7) &&
                                                     (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                    (**(code **)(plVar7[1] + 0x30))(plVar7);
                                                  }
                                                  if ((-1 < (int)*local_68) &&
                                                     (*local_68 = *local_68 + -1, *local_68 == 0)) {
                                                    (**(code **)(local_90[1] + 0x30))();
                                                  }
                                                  if ((-1 < (int)*local_80) &&
                                                     (*local_80 = *local_80 + -1, *local_80 == 0)) {
                                                    (**(code **)(local_88[1] + 0x30))();
                                                  }
                                                  if ((-1 < (int)*local_78) &&
                                                     (*local_78 = *local_78 + -1, *local_78 == 0)) {
                                                    (**(code **)(local_a8[1] + 0x30))();
                                                  }
                                                  if (((plVar8 != (longlong *)0x0) &&
                                                      (-1 < (int)*plVar8)) &&
                                                     (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                    (**(code **)(plVar8[1] + 0x30))(plVar8);
                                                  }
                                                  if (((plVar10 != (longlong *)0x0) &&
                                                      (-1 < (int)*plVar10)) &&
                                                     (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                                    (**(code **)(plVar10[1] + 0x30))(plVar10);
                                                  }
                                                  if (((local_res18 != (longlong *)0x0) &&
                                                      (-1 < (int)*local_res18)) &&
                                                     (*local_res18 = *local_res18 + -1,
                                                     *local_res18 == 0)) {
                                                    (**(code **)(local_res18[1] + 0x30))();
                                                  }
                                                  if (((local_c0 != (longlong *)0x0) &&
                                                      (-1 < (int)*local_c0)) &&
                                                     (*local_c0 = *local_c0 + -1, *local_c0 == 0)) {
                                                    (**(code **)(local_c0[1] + 0x30))();
                                                  }
                                                  if ((-1 < (int)*local_a0) &&
                                                     (*local_a0 = *local_a0 + -1, *local_a0 == 0)) {
                                                    (**(code **)(local_a0[1] + 0x30))();
                                                  }
                                                  if ((int)*local_70 < 0) {
                                                    return plVar7;
                                                  }
                                                  *local_70 = *local_70 + -1;
                                                  if (*local_70 != 0) {
                                                    return plVar7;
                                                  }
                                                  (**(code **)(local_70[1] + 0x30))();
                                                  return plVar7;
                                                }
                                                plVar18 = (longlong *)
                                                          PyUnicode_FromFormat
                                                                    (
                                                  "cannot access local variable \'%U\' where it is not associated with a value"
/* novels.pipeline.short_fiction_parsers#15="blocks" */
                                                  ,DAT_1803b6ca8);
                                                plVar7 = (longlong *)
                                                         FUN_180376770(param_1,*(undefined8 *)

                                                  PyExc_UnboundLocalError_exref,plVar18);
                                                if ((-1 < (int)*plVar18) &&
                                                   (*plVar18 = *plVar18 + -1, *plVar18 == 0)) {
                                                  (**(code **)(plVar18[1] + 0x30))(plVar18);
                                                }
                                                FUN_18000d1d0(param_1,plVar7);
                                                uVar17 = 0x45;
                                                goto LAB_18014dd8e;
                                              }
                                              plVar7 = (longlong *)(*pcVar12)(plVar18);
                                              if (plVar7 == (longlong *)0x0) goto LAB_18014dc0f;
                                              local_b8 = plVar7;
                                              if (((plVar9 != (longlong *)0x0) &&
                                                  (-1 < (int)*plVar9)) &&
                                                 (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                (**(code **)(plVar9[1] + 0x30))(plVar9);
                                              }
                                              iVar5 = (int)*plVar7 + 1;
                                              if (iVar5 != 0) {
                                                *(int *)plVar7 = iVar5;
                                              }
                                              local_d0 = plVar7;
                                              if (((plVar8 != (longlong *)0x0) &&
                                                  (-1 < (int)*plVar8)) &&
                                                 (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                (**(code **)(plVar8[1] + 0x30))();
                                              }
                                              *(undefined4 *)(plVar3 + 5) = 0x41;
                                              plVar7 = (longlong *)
/* novels.pipeline.short_fiction_parsers#10="end" */
                                                       FUN_18037e560(param_1,plVar7,DAT_1803b6c80);
                                              if (plVar7 == (longlong *)0x0) {
                                                plVar7 = *(longlong **)(param_1 + 0x60);
                                                uVar17 = 0x41;
                                                *(undefined8 *)(param_1 + 0x60) = 0;
                                                goto LAB_18014dc89;
                                              }
                                              local_c8 = plVar7;
                                              if (((plVar10 != (longlong *)0x0) &&
                                                  (-1 < (int)*plVar10)) &&
                                                 (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                                (**(code **)(plVar10[1] + 0x30))(plVar10);
                                              }
                                              plVar7 = (longlong *)
                                                       FUN_18014b800(param_1,plVar7,
                                                                     _Py_NoneStruct_exref);
                                              plVar8 = (longlong *)
                                                       FUN_180024e50(param_1,local_a0,plVar7);
                                              if ((-1 < (int)*plVar7) &&
                                                 (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                (**(code **)(plVar7[1] + 0x30))(plVar7);
                                              }
                                              if (plVar8 == (longlong *)0x0) {
LAB_18014dbb7:
                                                plVar7 = *(longlong **)(param_1 + 0x60);
                                                uVar17 = 0x42;
                                                *(undefined8 *)(param_1 + 0x60) = 0;
                                                goto LAB_18014dc89;
                                              }
                                              plVar7 = (longlong *)FUN_18037f0c0();
                                              if ((-1 < (int)*plVar8) &&
                                                 (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                (**(code **)(plVar8[1] + 0x30))(plVar8);
                                              }
                                              if (plVar7 == (longlong *)0x0) goto LAB_18014dbb7;
                                              *(undefined4 *)(plVar3 + 5) = 0x42;
                                              plVar8 = (longlong *)
/* novels.pipeline.short_fiction_parsers#12={"type": "tuple", "items": [{"tag": "w", "fileRange": [3998891, 3998893], "value": "\n"}]} */
                                                       FUN_1803770d0(param_1,plVar7,DAT_1803b6c90);
                                              if ((-1 < (int)*plVar7) &&
                                                 (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                (**(code **)(plVar7[1] + 0x30))(plVar7);
                                              }
                                              if (plVar8 == (longlong *)0x0) goto LAB_18014dbb7;
                                              if (((local_res18 != (longlong *)0x0) &&
                                                  (-1 < (int)*local_e0[0])) &&
                                                 (*local_e0[0] = *local_e0[0] + -1,
                                                 *local_e0[0] == 0)) {
                                                (**(code **)(local_res18[1] + 0x30))();
                                              }
                                              *(undefined4 *)(plVar3 + 5) = 0x43;
                                              plVar10 = (longlong *)
/* novels.pipeline.short_fiction_parsers#14="search" */
                                                        FUN_18037e6c0(param_1,local_a8,DAT_1803b6ca0
                                                                      ,plVar8);
                                              plVar7 = local_c0;
                                              local_res18 = plVar8;
                                              if (plVar10 == (longlong *)0x0) {
                                                plVar7 = *(longlong **)(param_1 + 0x60);
                                                uVar17 = 0x43;
                                                *(undefined8 *)(param_1 + 0x60) = 0;
                                                goto LAB_18014dc89;
                                              }
                                              plVar9 = plVar10;
                                              if (((local_c0 != (longlong *)0x0) &&
                                                  (-1 < (int)*local_c0)) &&
                                                 (*local_c0 = *local_c0 + -1, *local_c0 == 0)) {
                                                plVar9 = local_c0 + 1;
                                                local_c0 = plVar10;
                                                (**(code **)(*plVar9 + 0x30))(plVar7);
                                                plVar9 = local_c0;
                                              }
                                              local_c0 = plVar9;
                                              if (local_b0 == (longlong *)0x0) {
                                                plVar8 = (longlong *)
                                                         PyUnicode_FromFormat
                                                                   (
                                                  "cannot access local variable \'%U\' where it is not associated with a value"
/* novels.pipeline.short_fiction_parsers#15="blocks" */
                                                  ,DAT_1803b6ca8);
                                                plVar7 = (longlong *)
                                                         FUN_180376770(param_1,*(undefined8 *)

                                                  PyExc_UnboundLocalError_exref,plVar8);
                                                if ((-1 < (int)*plVar8) &&
                                                   (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                  (**(code **)(plVar8[1] + 0x30))(plVar8);
                                                }
                                                FUN_18000d1d0(param_1,plVar7);
                                                uVar17 = 0x44;
                                                goto LAB_18014dc89;
                                              }
                                              iVar5 = FUN_180001510(plVar10);
                                              if (iVar5 == -1) {
LAB_18014d98d:
                                                plVar7 = *(longlong **)(param_1 + 0x60);
                                                uVar17 = 0x44;
                                                *(undefined8 *)(param_1 + 0x60) = 0;
                                                goto LAB_18014dc89;
                                              }
                                              if (iVar5 != 0) {
                                                *(undefined4 *)(plVar3 + 5) = 0x44;
                                                plVar7 = (longlong *)
/* novels.pipeline.short_fiction_parsers#16="start" */
                                                         FUN_18037e560(param_1,plVar10,DAT_1803b6cb0
                                                                      );
                                                if (plVar7 != (longlong *)0x0) {
                                                  plVar10 = (longlong *)
                                                            FUN_18014b8e0(param_1,plVar7);
                                                  if ((-1 < (int)*plVar7) &&
                                                     (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                    (**(code **)(plVar7[1] + 0x30))(plVar7);
                                                  }
                                                  plVar8 = (longlong *)
                                                           FUN_180024e50(param_1,plVar8,plVar10);
                                                  if ((-1 < (int)*plVar10) &&
                                                     (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                                    (**(code **)(plVar10[1] + 0x30))(plVar10);
                                                  }
                                                  if (plVar8 != (longlong *)0x0) goto LAB_18014d9b0;
                                                }
                                                goto LAB_18014d98d;
                                              }
                                              iVar5 = (int)*plVar8 + 1;
                                              if (iVar5 != 0) {
                                                *(int *)plVar8 = iVar5;
                                              }
LAB_18014d9b0:
                                              lVar6 = plVar8[1];
                                              if (*(code **)(lVar6 + 0x90) == (code *)0x0) {
/* novels.pipeline.short_fiction_parsers#17="strip" */
                                                uVar14 = *(uint *)(DAT_1803b6cb8 + 0x20) & 0x60;
                                                if (*(code **)(lVar6 + 0x40) == (code *)0x0) {
/* novels.pipeline.short_fiction_parsers#17="strip" */
                                                  lVar16 = DAT_1803b6cb8 + 0x28;
                                                  if (uVar14 != 0x60) {
/* novels.pipeline.short_fiction_parsers#17="strip" */
                                                    lVar16 = *(longlong *)(DAT_1803b6cb8 + 0x30);
                                                  }
                                                  PyErr_Format(*(undefined8 *)
                                                                PyExc_AttributeError_exref,

                                                  "\'%s\' object has no attribute \'%s\'",
                                                  *(undefined8 *)(lVar6 + 0x18),lVar16);
                                                  plVar7 = (longlong *)0x0;
                                                }
                                                else {
/* novels.pipeline.short_fiction_parsers#17="strip" */
                                                  lVar16 = DAT_1803b6cb8 + 0x28;
                                                  if (uVar14 != 0x60) {
/* novels.pipeline.short_fiction_parsers#17="strip" */
                                                    lVar16 = *(longlong *)(DAT_1803b6cb8 + 0x30);
                                                  }
                                                  plVar7 = (longlong *)
                                                           (**(code **)(lVar6 + 0x40))
                                                                     (plVar8,lVar16);
                                                }
                                              }
                                              else {
                                                plVar7 = (longlong *)
                                                         (**(code **)(lVar6 + 0x90))
/* novels.pipeline.short_fiction_parsers#17="strip" */
                                                                   (plVar8,DAT_1803b6cb8);
                                              }
                                              if ((-1 < (int)*plVar8) &&
                                                 (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                (**(code **)(plVar8[1] + 0x30))(plVar8);
                                              }
                                              if (plVar7 == (longlong *)0x0) goto LAB_18014d98d;
                                              *(undefined4 *)(plVar3 + 5) = 0x44;
                                              lVar6 = FUN_180376210(param_1,plVar7);
                                              if ((-1 < (int)*plVar7) &&
                                                 (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                (**(code **)(plVar7[1] + 0x30))(plVar7);
                                              }
                                              plVar7 = local_b0;
                                              if (lVar6 == 0) goto LAB_18014d98d;
                                              lVar16 = local_b0[2];
                                              cVar4 = FUN_1803722b0(local_b0,lVar16 + 1);
                                              if (cVar4 != '\0') {
                                                *(longlong *)(plVar7[3] + lVar16 * 8) = lVar6;
                                              }
                                              lVar6 = *(longlong *)(param_1 + 0x10);
                                              if ((((*(int *)(*(longlong *)(lVar6 + 0x60) + 0x4d4)
                                                     != 0) || (*(int *)(lVar6 + 0x1a0) != 0)) &&
                                                  (iVar5 = Py_MakePendingCalls(), iVar5 < 0)) &&
                                                 (*(longlong *)(param_1 + 0x60) != 0))
                                              goto LAB_18014db32;
                                              if (*(int *)(lVar6 + 0x174) != 0) {
                                                PyEval_SaveThread();
                                                PyEval_AcquireThread(param_1);
                                              }
                                            } while (*(longlong *)(param_1 + 0x80) == 0);
                                            plVar7 = *(longlong **)(param_1 + 0x60);
                                            *(undefined8 *)(param_1 + 0x80) = 0;
                                            uVar11 = FUN_180373490(param_1);
                                            *(undefined8 *)(param_1 + 0x60) = uVar11;
                                            if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7))
                                               && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                              (**(code **)(plVar7[1] + 0x30))(plVar7);
                                            }
LAB_18014db32:
                                            plVar7 = *(longlong **)(param_1 + 0x60);
                                            *(undefined8 *)(param_1 + 0x60) = 0;
LAB_18014dc89:
                                            if ((-1 < (int)*local_b8) &&
                                               (*local_b8 = *local_b8 + -1, *local_b8 == 0)) {
                                              (**(code **)(local_b8[1] + 0x30))(local_b8);
                                            }
LAB_18014dca5:
                                            if ((-1 < (int)*plVar18) &&
                                               (*plVar18 = *plVar18 + -1, *plVar18 == 0)) {
                                              (**(code **)(plVar18[1] + 0x30))();
                                            }
                                          }
                                        }
                                      }
                                    }
                                    goto LAB_18014dd8e;
                                  }
/* novels.pipeline.short_fiction_parsers#0="re" */
                                  FUN_180373340(param_1,local_e0,DAT_1803b6c30);
                                  if ((-1 < (int)*plVar18) &&
                                     (*plVar18 = *plVar18 + -1, *plVar18 == 0)) {
                                    (**(code **)(plVar18[1] + 0x30))(plVar18);
                                  }
                                  plVar7 = local_e0[0];
                                  if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0))
                                  {
                                    lVar6 = plVar8[1];
                                    goto LAB_18014d469;
                                  }
                                }
                              }
LAB_18014d474:
                              uVar17 = 0x3e;
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
LAB_18014dd8e:
  piVar2 = (int *)plVar7[5];
  if (piVar2 == (int *)0x0) {
    pcVar12 = (code *)FUN_18038e2a0(plVar3,uVar17);
    plVar18 = (longlong *)plVar7[5];
    pcVar13 = pcVar12;
    if (pcVar12 == _Py_NoneStruct_exref) {
      pcVar13 = (code *)0x0;
    }
    else {
LAB_18014dde8:
      pcVar12 = pcVar13;
      if ((pcVar13 != (code *)0x0) && (*(int *)pcVar13 + 1 != 0)) {
        *(int *)pcVar13 = *(int *)pcVar13 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar2 + 6) == plVar3) goto LAB_18014de31;
    pcVar12 = (code *)FUN_18038e2a0(plVar3,uVar17);
    *(int **)(pcVar12 + 0x10) = piVar2;
    if (*piVar2 + 1 != 0) {
      *piVar2 = *piVar2 + 1;
    }
    plVar18 = (longlong *)plVar7[5];
    pcVar13 = pcVar12;
    if (pcVar12 != _Py_NoneStruct_exref) goto LAB_18014dde8;
    pcVar13 = (code *)0x0;
  }
  plVar7[5] = (longlong)pcVar13;
  if (((plVar18 != (longlong *)0x0) && (-1 < (int)*plVar18)) &&
     (*plVar18 = *plVar18 + -1, *plVar18 == 0)) {
    (**(code **)(plVar18[1] + 0x30))();
  }
  if (((pcVar12 != (code *)0x0) && (-1 < *(int *)pcVar12)) &&
     (*(longlong *)pcVar12 = *(longlong *)pcVar12 + -1, *(longlong *)pcVar12 == 0)) {
    (**(code **)(*(longlong *)(pcVar12 + 8) + 0x30))(pcVar12);
  }
LAB_18014de31:
  plVar10 = local_b0;
  plVar8 = local_c0;
  FUN_18036d0a0(plVar3,"oooooooooo",local_a0,local_70,local_b0,local_90,local_88,local_a8,local_d0,
                local_c8,local_res18,local_c0);
  plVar18 = DAT_1803b6b68;
  if (plVar3 == DAT_1803b6b68) {
    if ((-1 < (int)*DAT_1803b6b68) && (*DAT_1803b6b68 = *DAT_1803b6b68 + -1, *plVar18 == 0)) {
      (**(code **)(plVar18[1] + 0x30))();
    }
    DAT_1803b6b68 = (longlong *)0x0;
  }
  lVar6 = **(longlong **)(param_1 + 0x38);
  plVar3 = *(longlong **)(lVar6 + 0x30);
  **(longlong **)(param_1 + 0x38) = *(longlong *)(lVar6 + 8);
  *(undefined4 *)(plVar3 + 8) = 0xffffffff;
  plVar18 = (longlong *)plVar3[2];
  if (((plVar18 != (longlong *)0x0) && (plVar3[2] = 0, -1 < (int)*plVar18)) &&
     (*plVar18 = *plVar18 + -1, *plVar18 == 0)) {
    (**(code **)(plVar18[1] + 0x30))();
  }
  if ((-1 < (int)*plVar3) && (*plVar3 = *plVar3 + -1, *plVar3 == 0)) {
    (**(code **)(plVar3[1] + 0x30))(plVar3);
  }
  plVar3[10] = 0;
  if (((plVar10 != (longlong *)0x0) && (-1 < (int)*plVar10)) &&
     (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
    (**(code **)(plVar10[1] + 0x30))(plVar10);
  }
  if (((local_90 != (longlong *)0x0) && (-1 < (int)*local_90)) &&
     (*local_90 = *local_90 + -1, *local_90 == 0)) {
    (**(code **)(local_90[1] + 0x30))();
  }
  if (((local_88 != (longlong *)0x0) && (-1 < (int)*local_88)) &&
     (*local_88 = *local_88 + -1, *local_88 == 0)) {
    (**(code **)(local_88[1] + 0x30))();
  }
  if (((local_a8 != (longlong *)0x0) && (-1 < (int)*local_a8)) &&
     (*local_a8 = *local_a8 + -1, *local_a8 == 0)) {
    (**(code **)(local_a8[1] + 0x30))();
  }
  if (((local_d0 != (longlong *)0x0) && (-1 < (int)*local_d0)) &&
     (*local_d0 = *local_d0 + -1, *local_d0 == 0)) {
    (**(code **)(local_d0[1] + 0x30))();
  }
  if (((local_c8 != (longlong *)0x0) && (-1 < (int)*local_c8)) &&
     (*local_c8 = *local_c8 + -1, *local_c8 == 0)) {
    (**(code **)(local_c8[1] + 0x30))();
  }
  if (((local_res18 != (longlong *)0x0) && (-1 < (int)*local_res18)) &&
     (*local_res18 = *local_res18 + -1, *local_res18 == 0)) {
    (**(code **)(local_res18[1] + 0x30))(local_res18);
  }
  if (((plVar8 != (longlong *)0x0) && (-1 < (int)*plVar8)) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)
     ) {
    (**(code **)(plVar8[1] + 0x30))(plVar8);
  }
  if ((-1 < (int)*local_a0) && (*local_a0 = *local_a0 + -1, *local_a0 == 0)) {
    (**(code **)(local_a0[1] + 0x30))();
  }
  if ((-1 < (int)*local_70) && (*local_70 = *local_70 + -1, *local_70 == 0)) {
    (**(code **)(local_70[1] + 0x30))();
  }
  plVar3 = *(longlong **)(param_1 + 0x60);
  *(longlong **)(param_1 + 0x60) = plVar7;
  if (((plVar3 != (longlong *)0x0) && (-1 < (int)*plVar3)) && (*plVar3 = *plVar3 + -1, *plVar3 == 0)
     ) {
    (**(code **)(plVar3[1] + 0x30))();
  }
  return (longlong *)0x0;
}
