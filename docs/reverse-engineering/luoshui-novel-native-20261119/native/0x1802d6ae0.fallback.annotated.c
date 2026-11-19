/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x1802d6ae0 */

undefined8 FUN_1802d6ae0(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  int *piVar1;
  longlong *plVar2;
  char cVar3;
  int iVar4;
  uint uVar5;
  longlong *plVar6;
  longlong *plVar7;
  longlong *plVar8;
  longlong *plVar9;
  longlong *plVar10;
  longlong *plVar11;
  undefined8 uVar12;
  code *pcVar13;
  undefined8 uVar14;
  longlong lVar15;
  longlong lVar16;
  undefined4 uVar17;
  bool bVar18;
  longlong *local_res18;
  longlong local_res20;
  longlong *local_108;
  longlong *local_100;
  longlong *local_f8;
  longlong *local_f0;
  longlong *local_e8;
  longlong *local_e0;
  longlong *local_d0;
  longlong *local_c8;
  longlong *local_c0;
  longlong *local_b8;
  longlong *local_b0;
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
  longlong *local_58;
  longlong *local_50;
  longlong *local_48;
  
  plVar2 = DAT_1803c5480;
  local_90 = (longlong *)*param_3;
  plVar10 = (longlong *)0x0;
  plVar8 = (longlong *)param_3[1];
  local_98 = (longlong *)param_3[2];
  local_res20 = DAT_180391ea8;
  local_78 = (longlong *)0x0;
  local_70 = (longlong *)0x0;
  local_100 = (longlong *)0x0;
  local_e8 = (longlong *)0x0;
  local_f8 = (longlong *)0x0;
  local_c8 = (longlong *)0x0;
  local_c0 = (longlong *)0x0;
  local_b8 = (longlong *)0x0;
  local_a0 = (longlong *)0x0;
  local_a8 = (longlong *)0x0;
  local_68 = plVar8;
  if (DAT_1803c5480 == (longlong *)0x0) {
LAB_1802d6b8e:
    DAT_1803c5480 = (longlong *)FUN_18036cd10(param_1,DAT_1803c54b0,DAT_1803ca450,0x58);
  }
  else if ((1 < *DAT_1803c5480) || (DAT_1803c5480[2] != 0)) {
    if ((DAT_1803c5480 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803c5480 && (*DAT_1803c5480 = *DAT_1803c5480 + -1, *plVar2 == 0)))) {
      (**(code **)(plVar2[1] + 0x30))(plVar2);
    }
    goto LAB_1802d6b8e;
  }
  plVar2 = DAT_1803c5480;
  FUN_180001d00(param_1,DAT_1803c5480);
/* novels.services.state_reducer#3="get" */
  plVar6 = (longlong *)FUN_18037f0c0(param_1,plVar8,DAT_1803c54f8);
  if (plVar6 == (longlong *)0x0) {
    lVar16 = *(longlong *)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
    uVar17 = 0x43;
  }
  else {
    uVar17 = 0x43;
    *(undefined4 *)(plVar2 + 5) = 0x43;
    plVar7 = (longlong *)FUN_1803770d0(param_1,plVar6,DAT_1803c55c8);
    if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
      (**(code **)(plVar6[1] + 0x30))(plVar6);
    }
    if (plVar7 == (longlong *)0x0) {
      lVar16 = *(longlong *)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
    }
    else {
      iVar4 = PyObject_IsInstance(plVar7,PyDict_Type_exref);
      if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
        (**(code **)(plVar7[1] + 0x30))(plVar7);
      }
      if (iVar4 == -1) {
        lVar16 = *(longlong *)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
        plVar10 = (longlong *)0x0;
      }
      else if (iVar4 == 0) {
        local_d0 = (longlong *)FUN_180370ec0(param_1);
LAB_1802d6d2f:
        local_70 = local_d0;
/* novels.services.state_reducer#3="get" */
        plVar8 = (longlong *)FUN_18037f0c0(param_1,local_90,DAT_1803c54f8);
        if (plVar8 == (longlong *)0x0) {
          lVar16 = *(longlong *)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
        }
        else {
          *(undefined4 *)(plVar2 + 5) = 0x46;
          plVar6 = (longlong *)FUN_1803770d0(param_1,plVar8,DAT_1803c5520);
          if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
            (**(code **)(plVar8[1] + 0x30))(plVar8);
          }
          if (plVar6 == (longlong *)0x0) {
            lVar16 = *(longlong *)(param_1 + 0x60);
            *(undefined8 *)(param_1 + 0x60) = 0;
          }
          else {
            iVar4 = FUN_180001510(plVar6);
            if (iVar4 == -1) {
              local_res20 = *(longlong *)(param_1 + 0x60);
              *(undefined8 *)(param_1 + 0x60) = 0;
              lVar16 = local_res20;
              if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                (**(code **)(plVar6[1] + 0x30))(plVar6);
                lVar16 = local_res20;
              }
            }
            else {
              if (iVar4 != 1) {
                if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                  (**(code **)(plVar6[1] + 0x30))(plVar6);
                }
                plVar6 = (longlong *)FUN_180372160(param_1,0);
              }
              plVar8 = (longlong *)FUN_180001930(param_1,plVar6);
              if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                (**(code **)(plVar6[1] + 0x30))(plVar6);
              }
              if (plVar8 == (longlong *)0x0) {
                lVar16 = *(longlong *)(param_1 + 0x60);
                *(undefined8 *)(param_1 + 0x60) = 0;
              }
              else {
                plVar6 = (longlong *)FUN_180370ec0(param_1);
                local_f0 = plVar6;
                plVar9 = (longlong *)FUN_18000d430(plVar8);
                plVar7 = plVar10;
                if (plVar9 != (longlong *)0x0) {
LAB_1802d6e70:
                  plVar10 = plVar9;
                  if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) &&
                     (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                    (**(code **)(plVar7[1] + 0x30))(plVar7);
                  }
                  iVar4 = (int)*plVar10 + 1;
                  if (iVar4 != 0) {
                    *(int *)plVar10 = iVar4;
                  }
                  if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) &&
                     (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                    (**(code **)(plVar7[1] + 0x30))();
                  }
                  iVar4 = PyObject_IsInstance(plVar10,PyDict_Type_exref);
                  if (iVar4 == -1) goto LAB_1802d91f4;
                  bVar18 = iVar4 != 0;
                  if (iVar4 == 0) {
LAB_1802d7051:
                    if (!bVar18) {
LAB_1802d7189:
                      cVar3 = FUN_18000d5b0(param_1);
                      if (cVar3 == '\0') goto LAB_1802d91f4;
                      plVar9 = (longlong *)FUN_18000d430(plVar8);
                      plVar6 = local_f0;
                      plVar7 = plVar10;
                      if (plVar9 == (longlong *)0x0) goto LAB_1802d7272;
                      goto LAB_1802d6e70;
                    }
/* novels.services.state_reducer#3="get" */
                    plVar7 = (longlong *)FUN_18037f0c0(param_1,plVar10,DAT_1803c54f8);
                    if (plVar7 != (longlong *)0x0) {
                      *(undefined4 *)(plVar2 + 5) = 0x45;
                      plVar6 = (longlong *)FUN_1803770d0(param_1,plVar7,DAT_1803c55d0);
                      if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                        (**(code **)(plVar7[1] + 0x30))(plVar7);
                      }
                      if (plVar6 == (longlong *)0x0) goto LAB_1802d71ed;
                      iVar4 = FUN_180001510(plVar6);
                      if (iVar4 == -1) {
                        local_res20 = *(longlong *)(param_1 + 0x60);
                        *(undefined8 *)(param_1 + 0x60) = 0;
LAB_1802d71c1:
                        if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                          (**(code **)(plVar6[1] + 0x30))(plVar6);
                        }
                        goto LAB_1802d9207;
                      }
                      if (iVar4 != 1) {
                        if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                          (**(code **)(plVar6[1] + 0x30))(plVar6);
                        }
/* novels.services.state_reducer#14="" */
                        plVar6 = DAT_1803c5550;
/* novels.services.state_reducer#14="" */
                        iVar4 = (int)*DAT_1803c5550 + 1;
                        if (iVar4 != 0) {
/* novels.services.state_reducer#14="" */
                          *(int *)DAT_1803c5550 = iVar4;
                        }
                      }
                      plVar7 = (longlong *)FUN_1803749a0(plVar6);
                      if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                        (**(code **)(plVar6[1] + 0x30))(plVar6);
                      }
                      if (plVar7 != (longlong *)0x0) {
                        plVar6 = (longlong *)FUN_180370dc0(param_1,plVar10,0);
                        if (plVar6 != (longlong *)0x0) {
                          iVar4 = PyDict_SetItem(local_f0,plVar7,plVar6);
                          if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                            (**(code **)(plVar6[1] + 0x30))(plVar6);
                          }
                          if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                            (**(code **)(plVar7[1] + 0x30))(plVar7);
                          }
                          if (iVar4 == 0) goto LAB_1802d7189;
                          goto LAB_1802d91f4;
                        }
                        local_res20 = *(longlong *)(param_1 + 0x60);
                        *(undefined8 *)(param_1 + 0x60) = 0;
                        if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                          (**(code **)(plVar7[1] + 0x30))(plVar7);
                        }
                        goto LAB_1802d9207;
                      }
                    }
                  }
                  else {
/* novels.services.state_reducer#3="get" */
                    plVar6 = (longlong *)FUN_18037f0c0(param_1,plVar10,DAT_1803c54f8);
                    if (plVar6 != (longlong *)0x0) {
                      *(undefined4 *)(plVar2 + 5) = 0x47;
                      plVar7 = (longlong *)FUN_1803770d0(param_1,plVar6,DAT_1803c55d0);
                      if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                        (**(code **)(plVar6[1] + 0x30))(plVar6);
                      }
                      if (plVar7 != (longlong *)0x0) {
                        iVar4 = FUN_180001510(plVar7);
                        if (iVar4 == -1) {
                          local_res20 = *(longlong *)(param_1 + 0x60);
                          *(undefined8 *)(param_1 + 0x60) = 0;
                          if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                            (**(code **)(plVar7[1] + 0x30))(plVar7);
                          }
                          goto LAB_1802d9207;
                        }
                        if (iVar4 != 1) {
                          if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                            (**(code **)(plVar7[1] + 0x30))(plVar7);
                          }
/* novels.services.state_reducer#14="" */
                          plVar7 = DAT_1803c5550;
/* novels.services.state_reducer#14="" */
                          iVar4 = (int)*DAT_1803c5550 + 1;
                          if (iVar4 != 0) {
/* novels.services.state_reducer#14="" */
                            *(int *)DAT_1803c5550 = iVar4;
                          }
                        }
                        plVar6 = (longlong *)FUN_1803749a0(plVar7);
                        if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                          (**(code **)(plVar7[1] + 0x30))(plVar7);
                        }
                        if (plVar6 != (longlong *)0x0) {
/* novels.services.state_reducer#31="strip" */
                          plVar7 = (longlong *)FUN_18037f0c0(param_1,plVar6,DAT_1803c55d8);
                          if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                            (**(code **)(plVar6[1] + 0x30))(plVar6);
                          }
                          if (plVar7 != (longlong *)0x0) {
                            *(undefined4 *)(plVar2 + 5) = 0x47;
                            plVar6 = (longlong *)FUN_180376210(param_1,plVar7);
                            if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                              (**(code **)(plVar7[1] + 0x30))(plVar7);
                            }
                            if (plVar6 != (longlong *)0x0) {
                              iVar4 = FUN_180001510(plVar6);
                              if (iVar4 != -1) {
                                bVar18 = iVar4 != 0;
                                if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                  (**(code **)(plVar6[1] + 0x30))(plVar6);
                                }
                                goto LAB_1802d7051;
                              }
                              local_res20 = *(longlong *)(param_1 + 0x60);
                              *(undefined8 *)(param_1 + 0x60) = 0;
                              goto LAB_1802d71c1;
                            }
                          }
                          goto LAB_1802d91f4;
                        }
                      }
                    }
                  }
LAB_1802d71ed:
                  local_res20 = *(longlong *)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                  goto LAB_1802d9207;
                }
LAB_1802d7272:
                cVar3 = FUN_1800247e0(param_1,*(undefined8 *)PyExc_StopIteration_exref);
                if (cVar3 != '\0') {
                  iVar4 = (int)*plVar6 + 1;
                  if (iVar4 != 0) {
                    *(int *)plVar6 = iVar4;
                  }
                  if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                    (**(code **)(plVar8[1] + 0x30))(plVar8);
                  }
                  if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                    (**(code **)(plVar6[1] + 0x30))(plVar6);
                  }
                  if ((plVar10 != (longlong *)0x0) && (-1 < (int)*plVar10)) {
                    *plVar10 = *plVar10 + -1;
                    iVar4 = (int)*plVar10;
                    if (*plVar10 == 0) {
                      (**(code **)(plVar10[1] + 0x30))(plVar10);
                      iVar4 = (int)*plVar10;
                    }
                    if ((-1 < iVar4) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                      (**(code **)(plVar10[1] + 0x30))(plVar10);
                    }
                  }
                  plVar8 = local_d0;
                  local_78 = plVar6;
/* novels.services.state_reducer#3="get" */
                  plVar10 = (longlong *)FUN_18037f0c0(param_1,local_d0,DAT_1803c54f8);
                  uVar17 = 0x4a;
                  if (plVar10 == (longlong *)0x0) {
                    lVar16 = *(longlong *)(param_1 + 0x60);
                    *(undefined8 *)(param_1 + 0x60) = 0;
                    plVar10 = local_100;
                  }
                  else {
                    local_res18 = (longlong *)CONCAT44(local_res18._4_4_,0x4a);
                    *(undefined4 *)(plVar2 + 5) = 0x4a;
                    plVar6 = (longlong *)FUN_1803770d0(param_1,plVar10,DAT_1803c55e0);
                    if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                      (**(code **)(plVar10[1] + 0x30))(plVar10);
                    }
                    if (plVar6 == (longlong *)0x0) {
                      FUN_18000ce30(param_1,&local_res20);
                      plVar10 = local_100;
                      lVar16 = local_res20;
                    }
                    else {
                      iVar4 = FUN_180001510(plVar6);
                      if (iVar4 == -1) {
                        FUN_18000ce30(param_1,&local_res20);
                        plVar10 = local_100;
                        lVar16 = local_res20;
                        if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                          (**(code **)(plVar6[1] + 0x30))(plVar6);
                          plVar10 = local_100;
                          lVar16 = local_res20;
                        }
                      }
                      else {
                        if (iVar4 != 1) {
                          if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                            (**(code **)(plVar6[1] + 0x30))(plVar6);
                          }
                          plVar6 = (longlong *)FUN_180372160(param_1,0);
                        }
                        plVar7 = (longlong *)FUN_180001930(param_1,plVar6);
                        local_108 = plVar7;
                        if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                          (**(code **)(plVar6[1] + 0x30))(plVar6);
                        }
                        if (plVar7 == (longlong *)0x0) {
                          FUN_18000ce30(param_1,&local_res20);
                          plVar10 = local_100;
                          lVar16 = local_res20;
                        }
                        else {
                          plVar6 = (longlong *)FUN_18000d430(plVar7);
                          plVar10 = local_b8;
                          local_b0 = plVar6;
                          if (plVar6 != (longlong *)0x0) {
                            do {
                              plVar11 = local_c8;
                              plVar9 = local_e8;
                              plVar7 = local_f8;
                              plVar8 = local_100;
                              local_b8 = plVar6;
                              local_b0 = plVar6;
                              if (((plVar10 != (longlong *)0x0) && (-1 < (int)*plVar10)) &&
                                 (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                (**(code **)(plVar10[1] + 0x30))();
                              }
                              iVar4 = (int)*plVar6 + 1;
                              if (iVar4 != 0) {
                                *(int *)plVar6 = iVar4;
                              }
                              local_100 = plVar6;
                              if (((plVar8 != (longlong *)0x0) && (-1 < (int)*plVar8)) &&
                                 (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                (**(code **)(plVar8[1] + 0x30))();
                              }
                              iVar4 = PyObject_IsInstance(plVar6,PyDict_Type_exref);
                              if (iVar4 == -1) {
LAB_1802d818d:
                                FUN_18000ce30(param_1,&local_res20);
                                uVar17 = 0x4b;
                                goto LAB_1802d9198;
                              }
                              pcVar13 = _Py_FalseStruct_exref;
                              if (iVar4 != 0) {
                                pcVar13 = _Py_TrueStruct_exref;
                              }
                              iVar4 = FUN_180001510(pcVar13);
                              if (iVar4 == -1) goto LAB_1802d818d;
                              plVar10 = plVar6;
                              if (iVar4 == 0) goto LAB_1802d7c28;
/* novels.services.state_reducer#3="get" */
                              plVar8 = (longlong *)FUN_18037f0c0(param_1,plVar6,DAT_1803c54f8);
                              if (plVar8 == (longlong *)0x0) goto LAB_1802d8140;
                              *(undefined4 *)(plVar2 + 5) = 0x4d;
                              plVar6 = (longlong *)FUN_1803770d0(param_1,plVar8,DAT_1803c55d0);
                              if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                (**(code **)(plVar8[1] + 0x30))(plVar8);
                              }
                              if (plVar6 == (longlong *)0x0) {
LAB_1802d8126:
                                FUN_18000ce30(param_1,&local_res20);
                                uVar17 = 0x4d;
                                goto LAB_1802d9198;
                              }
                              iVar4 = FUN_180001510(plVar6);
                              if (iVar4 == -1) {
                                FUN_18000ce30(param_1,&local_res20);
                                if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                  lVar16 = FUN_18000cc40(plVar6);
                                  (**(code **)(lVar16 + 0x30))(plVar6);
                                }
                                uVar17 = 0x4d;
                                goto LAB_1802d9198;
                              }
                              if (iVar4 != 1) {
                                if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                  (**(code **)(plVar6[1] + 0x30))(plVar6);
                                }
/* novels.services.state_reducer#14="" */
                                plVar6 = DAT_1803c5550;
/* novels.services.state_reducer#14="" */
                                iVar4 = (int)*DAT_1803c5550 + 1;
                                if (iVar4 != 0) {
/* novels.services.state_reducer#14="" */
                                  *(int *)DAT_1803c5550 = iVar4;
                                }
                              }
                              plVar8 = (longlong *)FUN_1803749a0(plVar6);
                              if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                (**(code **)(plVar6[1] + 0x30))(plVar6);
                              }
                              if (plVar8 == (longlong *)0x0) {
LAB_1802d8140:
                                FUN_18000ce30(param_1,&local_res20);
                                uVar17 = 0x4d;
                                goto LAB_1802d9198;
                              }
/* novels.services.state_reducer#31="strip" */
                              plVar6 = (longlong *)FUN_18037f0c0(param_1,plVar8,DAT_1803c55d8);
                              if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                (**(code **)(plVar8[1] + 0x30))(plVar8);
                              }
                              if (plVar6 == (longlong *)0x0) goto LAB_1802d8140;
                              *(undefined4 *)(plVar2 + 5) = 0x4d;
                              plVar8 = (longlong *)FUN_180376210(param_1,plVar6);
                              local_e0 = plVar8;
                              if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                (**(code **)(plVar6[1] + 0x30))(plVar6);
                              }
                              if (plVar8 == (longlong *)0x0) goto LAB_1802d8126;
                              local_e8 = plVar8;
                              if (((plVar9 != (longlong *)0x0) && (-1 < (int)*plVar9)) &&
                                 (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                (**(code **)(plVar9[1] + 0x30))(plVar9);
                              }
                              iVar4 = FUN_180001510(plVar8);
                              plVar6 = local_f0;
                              if (iVar4 == -1) {
                                FUN_18000ce30(param_1,&local_res20);
                                uVar17 = 0x4e;
                                goto LAB_1802d9198;
                              }
                              if (iVar4 == 0) goto LAB_1802d7c28;
                              if (local_f0 == (longlong *)0x0) {
/* novels.services.state_reducer#33="hooks_by_id" */
                                FUN_180370fe0(param_1,&local_res20,DAT_1803c55e8);
                                FUN_18000d1d0(param_1,local_res20);
                                uVar17 = 0x50;
                                goto LAB_1802d9198;
                              }
/* novels.services.state_reducer#3="get" */
                              plVar10 = (longlong *)FUN_18037f0c0(param_1,local_f0,DAT_1803c54f8);
                              if (plVar10 == (longlong *)0x0) {
                                FUN_18000ce30(param_1,&local_res20);
                                uVar17 = 0x50;
                                goto LAB_1802d9198;
                              }
                              *(undefined4 *)(plVar2 + 5) = 0x50;
                              plVar9 = (longlong *)FUN_180376770(param_1,plVar10,plVar8);
                              if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0))
                              {
                                (**(code **)(plVar10[1] + 0x30))(plVar10);
                              }
                              if (plVar9 == (longlong *)0x0) {
                                FUN_18000ce30(param_1,&local_res20);
                                uVar17 = 0x50;
                                goto LAB_1802d9198;
                              }
                              local_f8 = plVar9;
                              if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) &&
                                 (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                (**(code **)(plVar7[1] + 0x30))(plVar7);
                              }
                              iVar4 = FUN_180001510(plVar9);
                              if (iVar4 == -1) {
                                FUN_18000ce30(param_1,&local_res20);
                                uVar17 = 0x51;
                                goto LAB_1802d9198;
                              }
                              if (iVar4 != 0) {
                                lVar16 = FUN_1802d3640();
                                if (lVar16 == 0) {
/* novels.services.state_reducer#34="_merge_hook_record" */
                                  FUN_180373340(param_1,&local_res20,DAT_1803c55f0);
                                  uVar17 = 0x52;
                                }
                                else {
                                  *(undefined4 *)(plVar2 + 5) = 0x52;
                                  local_58 = local_b0;
                                  local_60 = plVar9;
                                  plVar10 = (longlong *)FUN_180377750(param_1,lVar16,&local_60);
                                  if (plVar10 == (longlong *)0x0) {
                                    FUN_18000ce30(param_1,&local_res20);
                                    uVar17 = 0x52;
                                  }
                                  else {
                                    cVar3 = FUN_18000d470(param_1,plVar6,plVar8,plVar10);
                                    if ((-1 < (int)*plVar10) &&
                                       (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                      (**(code **)(plVar10[1] + 0x30))(plVar10);
                                    }
                                    plVar10 = local_b8;
                                    if (cVar3 != '\0') goto LAB_1802d7c28;
                                    FUN_18000ce30(param_1,&local_res20);
                                    uVar17 = 0x52;
                                  }
                                }
                                goto LAB_1802d9198;
                              }
                              lVar16 = FUN_1802d3420();
                              if (lVar16 == 0) {
/* novels.services.state_reducer#35="_evaluate_hook_admission" */
                                FUN_180373340(param_1,&local_res20,DAT_1803c55f8);
                                uVar17 = 0x54;
                                goto LAB_1802d9198;
                              }
/* novels.services.state_reducer#36="values" */
                              plVar8 = (longlong *)FUN_18037f0c0(param_1,local_f0,DAT_1803c5600);
                              if (plVar8 == (longlong *)0x0) {
LAB_1802d8062:
                                FUN_18000ce30(param_1,&local_res20);
                                uVar17 = 0x54;
                                goto LAB_1802d9198;
                              }
                              *(undefined4 *)(plVar2 + 5) = 0x54;
                              plVar10 = (longlong *)FUN_180376210(param_1,plVar8);
                              if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                (**(code **)(plVar8[1] + 0x30))(plVar8);
                              }
                              if (plVar10 == (longlong *)0x0) goto LAB_1802d8062;
                              plVar8 = (longlong *)FUN_180372690(param_1,plVar10);
                              if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0))
                              {
                                (**(code **)(plVar10[1] + 0x30))(plVar10);
                              }
                              if (plVar8 == (longlong *)0x0) goto LAB_1802d8062;
                              local_50 = local_b0;
                              *(undefined4 *)(plVar2 + 5) = 0x54;
                              local_48 = plVar8;
                              plVar10 = (longlong *)FUN_180377750(param_1,lVar16,&local_50);
                              if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                (**(code **)(plVar8[1] + 0x30))(plVar8);
                              }
                              if (plVar10 == (longlong *)0x0) {
                                FUN_18000ce30(param_1,&local_res20);
                                uVar17 = 0x54;
                                goto LAB_1802d9198;
                              }
                              local_c8 = plVar10;
                              if (((plVar11 != (longlong *)0x0) && (-1 < (int)*plVar11)) &&
                                 (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                (**(code **)(plVar11[1] + 0x30))(plVar11);
                              }
/* novels.services.state_reducer#37="admit" */
                              plVar8 = (longlong *)FUN_180024e50(param_1,plVar10,DAT_1803c5608);
                              if (plVar8 == (longlong *)0x0) {
                                FUN_18000ce30(param_1,&local_res20);
                                uVar17 = 0x55;
                                goto LAB_1802d9198;
                              }
                              iVar4 = FUN_180001510(plVar8);
                              if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                (**(code **)(plVar8[1] + 0x30))(plVar8);
                              }
                              if (iVar4 == -1) {
LAB_1802d7d13:
                                FUN_18000ce30(param_1,&local_res20);
                                uVar17 = 0x55;
                                goto LAB_1802d9198;
                              }
                              uVar5 = (uint)(iVar4 == 0);
                              if (iVar4 == 0) {
/* novels.services.state_reducer#38="reason" */
                                plVar8 = (longlong *)FUN_180024e50(param_1,plVar10,DAT_1803c5610);
                                if (plVar8 != (longlong *)0x0) {
/* novels.services.state_reducer#39="duplicate_family" */
                                  uVar5 = FUN_1803836d0(plVar8,DAT_1803c5618);
                                  if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0))
                                  {
                                    lVar16 = FUN_18000cc40(plVar8);
                                    (**(code **)(lVar16 + 0x30))(plVar8);
                                  }
                                  if (uVar5 != 0xffffffff) goto LAB_1802d7979;
                                }
                                goto LAB_1802d7d13;
                              }
LAB_1802d7979:
                              if (uVar5 != 1) {
                                plVar8 = (longlong *)FUN_180370dc0(param_1,local_b0,0);
                                if (plVar8 != (longlong *)0x0) {
                                  cVar3 = FUN_18000d470(param_1,local_f0,local_e0,plVar8);
                                  if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0))
                                  {
                                    lVar16 = FUN_18000cc40(plVar8);
                                    (**(code **)(lVar16 + 0x30))(plVar8);
                                  }
                                  if (cVar3 != '\0') {
                                    cVar3 = FUN_18000d5b0();
                                    plVar10 = local_b8;
                                    if (cVar3 != '\0') goto LAB_1802d7c28;
                                    FUN_18000ce30(param_1,&local_res20);
                                    uVar17 = local_res18._0_4_;
                                    goto LAB_1802d9198;
                                  }
                                }
                                FUN_18000ce30(param_1,&local_res20);
                                uVar17 = 0x5e;
                                goto LAB_1802d9198;
                              }
/* novels.services.state_reducer#3="get" */
                              plVar8 = (longlong *)FUN_18037f0c0(param_1,plVar10,DAT_1803c54f8);
                              if (plVar8 == (longlong *)0x0) {
                                FUN_18000ce30(param_1,&local_res20);
                                uVar17 = 0x56;
                                goto LAB_1802d9198;
                              }
                              *(undefined4 *)(plVar2 + 5) = 0x56;
                              plVar10 = (longlong *)FUN_1803770d0(param_1,plVar8,DAT_1803c5620);
                              if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                lVar16 = FUN_18000cc40(plVar8);
                                (**(code **)(lVar16 + 0x30))(plVar8);
                              }
                              if (plVar10 == (longlong *)0x0) {
                                FUN_18000ce30(param_1,&local_res20);
                                uVar17 = 0x56;
                                goto LAB_1802d9198;
                              }
                              plVar8 = plVar10;
                              if (((local_c0 != (longlong *)0x0) && (-1 < (int)*local_c0)) &&
                                 (*local_c0 = *local_c0 + -1, *local_c0 == 0)) {
                                plVar8 = local_c0 + 1;
                                local_c0 = plVar10;
                                (**(code **)(*plVar8 + 0x30))();
                                plVar8 = local_c0;
                              }
                              local_c0 = plVar8;
                              plVar8 = local_f0;
/* novels.services.state_reducer#3="get" */
                              plVar6 = (longlong *)FUN_18037f0c0(param_1,local_f0,DAT_1803c54f8);
                              if (plVar6 == (longlong *)0x0) {
                                FUN_18000ce30(param_1,&local_res20);
                                uVar17 = 0x57;
                                goto LAB_1802d9198;
                              }
                              iVar4 = FUN_180001510(plVar10);
                              if (iVar4 == -1) {
                                FUN_18000ce30(param_1,&local_res20);
                                if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                  lVar16 = FUN_18000cc40(plVar6);
                                  (**(code **)(lVar16 + 0x30))(plVar6);
                                }
                                uVar17 = 0x57;
                                goto LAB_1802d9198;
                              }
/* novels.services.state_reducer#14="" */
                              plVar7 = DAT_1803c5550;
                              if (iVar4 == 1) {
                                plVar7 = plVar10;
                              }
                              plVar7 = (longlong *)FUN_1803749a0(plVar7);
                              if (plVar7 == (longlong *)0x0) {
                                FUN_18000ce30(param_1,&local_res20);
                                if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                  lVar16 = FUN_18000cc40(plVar6);
                                  (**(code **)(lVar16 + 0x30))(plVar6);
                                }
                                uVar17 = 0x57;
                                goto LAB_1802d9198;
                              }
                              *(undefined4 *)(plVar2 + 5) = 0x57;
                              plVar11 = (longlong *)FUN_180376770(param_1,plVar6,plVar7);
                              if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                lVar16 = FUN_18000cc40(plVar6);
                                (**(code **)(lVar16 + 0x30))(plVar6);
                              }
                              if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                lVar16 = FUN_18000cc40(plVar7);
                                (**(code **)(lVar16 + 0x30))(plVar7);
                              }
                              if (plVar11 == (longlong *)0x0) {
                                FUN_18000ce30(param_1,&local_res20);
                                uVar17 = 0x57;
                                goto LAB_1802d9198;
                              }
                              local_f8 = plVar11;
                              if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                lVar16 = FUN_18000cc40(plVar9);
                                (**(code **)(lVar16 + 0x30))(plVar9);
                              }
                              iVar4 = FUN_180001510(plVar11);
/* novels.services.state_reducer#41="duplicate active hook family: " */
                              piVar1 = DAT_1803c5628;
                              if (iVar4 == -1) {
                                FUN_18000ce30(param_1,&local_res20);
                                uVar17 = 0x58;
                                goto LAB_1802d9198;
                              }
                              if (iVar4 == 0) {
                                plVar8 = (longlong *)FUN_180374a40(param_1,4);
                                if (*piVar1 + 1 != 0) {
                                  *piVar1 = *piVar1 + 1;
                                }
                                FUN_18000cc80(plVar8,0,piVar1);
/* novels.services.state_reducer#14="" */
                                lVar16 = FUN_180375520(param_1,local_e0,DAT_1803c5550);
                                if (lVar16 != 0) {
                                  FUN_18000cc80(plVar8,1,lVar16);
/* novels.services.state_reducer#42=" overlaps " */
                                  if (*DAT_1803c5630 + 1 != 0) {
/* novels.services.state_reducer#42=" overlaps " */
                                    *DAT_1803c5630 = *DAT_1803c5630 + 1;
                                  }
                                  FUN_18000cc80(plVar8,2);
/* novels.services.state_reducer#14="" */
                                  lVar16 = FUN_180375520(param_1,plVar10,DAT_1803c5550);
                                  if (lVar16 != 0) {
                                    FUN_18000cc80(plVar8,3,lVar16);
/* novels.services.state_reducer#14="" */
                                    plVar10 = (longlong *)PyUnicode_Join(DAT_1803c5550,plVar8);
                                    if ((-1 < (int)*plVar8) &&
                                       (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                      lVar16 = FUN_18000cc40(plVar8);
                                      (**(code **)(lVar16 + 0x30))(plVar8);
                                    }
                                    pcVar13 = PyExc_ValueError_exref;
                                    if (plVar10 == (longlong *)0x0) {
                                      FUN_18000ce30(param_1,&local_res20);
                                      uVar17 = 0x5a;
                                    }
                                    else {
                                      *(undefined4 *)(plVar2 + 5) = 0x59;
                                      lVar16 = FUN_180376770(param_1,*(undefined8 *)pcVar13,plVar10)
                                      ;
                                      if ((-1 < (int)*plVar10) &&
                                         (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                        lVar15 = FUN_18000cc40(plVar10);
                                        (**(code **)(lVar15 + 0x30))(plVar10);
                                      }
                                      local_res20 = lVar16;
                                      FUN_180373100(param_1,&local_res20);
                                      uVar17 = 0x59;
                                    }
                                    goto LAB_1802d9198;
                                  }
                                }
                                uVar17 = 0x5a;
                                FUN_18000ce30(param_1,&local_res20);
                                if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                  lVar16 = FUN_18000cc40(plVar8);
                                  (**(code **)(lVar16 + 0x30))(plVar8);
                                }
                                goto LAB_1802d9198;
                              }
                              lVar16 = FUN_1802d3640();
                              if (lVar16 == 0) {
/* novels.services.state_reducer#34="_merge_hook_record" */
                                FUN_180373340(param_1,&local_res20,DAT_1803c55f0);
                                uVar17 = 0x5c;
                                goto LAB_1802d9198;
                              }
                              local_80 = local_b0;
                              *(undefined4 *)(plVar2 + 5) = 0x5c;
                              local_88 = plVar11;
                              plVar10 = (longlong *)FUN_180377750(param_1,lVar16,&local_88);
                              if (plVar10 == (longlong *)0x0) {
                                FUN_18000ce30(param_1,&local_res20);
                                uVar17 = 0x5c;
                                goto LAB_1802d9198;
                              }
/* novels.services.state_reducer#43="hookId" */
                              plVar6 = (longlong *)FUN_180024e50(param_1,plVar11,DAT_1803c5638);
                              if (plVar6 == (longlong *)0x0) {
                                FUN_18000ce30(param_1,&local_res20);
                                if ((-1 < (int)*plVar10) &&
                                   (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                  lVar16 = FUN_18000cc40(plVar10);
                                  (**(code **)(lVar16 + 0x30))(plVar10);
                                }
                                uVar17 = 0x5c;
                                goto LAB_1802d9198;
                              }
                              cVar3 = FUN_18000d470(param_1,plVar8,plVar6,plVar10);
                              if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                lVar16 = FUN_18000cc40(plVar6);
                                (**(code **)(lVar16 + 0x30))(plVar6);
                              }
                              if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0))
                              {
                                lVar16 = FUN_18000cc40(plVar10);
                                (**(code **)(lVar16 + 0x30))(plVar10);
                              }
                              plVar10 = local_b8;
                              if (cVar3 == '\0') {
                                FUN_18000ce30(param_1,&local_res20);
                                uVar17 = 0x5c;
                                goto LAB_1802d9198;
                              }
LAB_1802d7c28:
                              plVar7 = local_108;
                              plVar6 = (longlong *)FUN_18000d430(local_108);
                            } while (plVar6 != (longlong *)0x0);
                            plVar8 = local_d0;
                            uVar17 = local_res18._0_4_;
                            local_b0 = plVar6;
                          }
                          cVar3 = FUN_1800012a0(param_1);
                          plVar10 = local_100;
                          if (cVar3 == '\0') {
                            FUN_18000ce30(param_1,&local_res20);
LAB_1802d9198:
                            lVar16 = local_res20;
                            FUN_18000ce70(&local_res20);
                            plVar10 = local_100;
                            if (((local_100 != (longlong *)0x0) && (-1 < (int)*local_100)) &&
                               (*local_100 = *local_100 + -1, *local_100 == 0)) {
                              (**(code **)(local_100[1] + 0x30))(local_100);
                            }
                            if ((-1 < (int)*local_108) &&
                               (*local_108 = *local_108 + -1, *local_108 == 0)) {
                              (**(code **)(local_108[1] + 0x30))();
                            }
                          }
                          else {
                            if (((local_100 != (longlong *)0x0) && (-1 < (int)*local_100)) &&
                               (*local_100 = *local_100 + -1, *local_100 == 0)) {
                              lVar16 = FUN_18000cc40(local_100);
                              (**(code **)(lVar16 + 0x30))(plVar10);
                            }
                            if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                              lVar16 = FUN_18000cc40(plVar7);
                              (**(code **)(lVar16 + 0x30))(plVar7);
                            }
/* novels.services.state_reducer#3="get" */
                            plVar8 = (longlong *)FUN_18037f0c0(param_1,plVar8,DAT_1803c54f8);
                            if (plVar8 == (longlong *)0x0) {
                              FUN_18000ce30(param_1,&local_res20);
                              uVar17 = 0x60;
                              lVar16 = local_res20;
                            }
                            else {
                              uVar17 = 0x60;
                              local_res18 = (longlong *)CONCAT44(local_res18._4_4_,0x60);
                              *(undefined4 *)(plVar2 + 5) = 0x60;
                              plVar6 = (longlong *)FUN_1803770d0(param_1,plVar8,DAT_1803c5640);
                              if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                lVar16 = FUN_18000cc40(plVar8);
                                (**(code **)(lVar16 + 0x30))(plVar8);
                              }
                              if (plVar6 == (longlong *)0x0) {
                                FUN_18000ce30(param_1,&local_res20);
                                lVar16 = local_res20;
                              }
                              else {
                                iVar4 = FUN_180001510(plVar6);
                                if (iVar4 == -1) {
                                  FUN_18000ce30(param_1,&local_res20);
                                  lVar16 = local_res20;
                                  if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0))
                                  {
                                    lVar16 = FUN_18000cc40(plVar6);
                                    (**(code **)(lVar16 + 0x30))(plVar6);
                                    lVar16 = local_res20;
                                  }
                                }
                                else {
                                  if (iVar4 != 1) {
                                    if ((-1 < (int)*plVar6) &&
                                       (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                      lVar16 = FUN_18000cc40(plVar6);
                                      (**(code **)(lVar16 + 0x30))(plVar6);
                                    }
                                    plVar6 = (longlong *)FUN_180372160(param_1,0);
                                  }
                                  plVar8 = (longlong *)FUN_180001930(param_1,plVar6);
                                  if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0))
                                  {
                                    lVar16 = FUN_18000cc40(plVar6);
                                    (**(code **)(lVar16 + 0x30))(plVar6);
                                  }
                                  if (plVar8 == (longlong *)0x0) {
                                    FUN_18000ce30(param_1,&local_res20);
                                    plVar10 = local_100;
                                    lVar16 = local_res20;
                                  }
                                  else {
                                    plVar10 = (longlong *)FUN_18000d430(plVar8);
                                    if (plVar10 != (longlong *)0x0) {
                                      do {
                                        plVar9 = local_a0;
                                        plVar7 = local_e8;
                                        plVar6 = local_f8;
                                        plVar11 = plVar10;
                                        if (((local_a0 != (longlong *)0x0) && (-1 < (int)*local_a0))
                                           && (*local_a0 = *local_a0 + -1, *local_a0 == 0)) {
                                          local_a0 = plVar10;
                                          lVar16 = FUN_18000cc40();
                                          (**(code **)(lVar16 + 0x30))(plVar9);
                                          plVar11 = local_a0;
                                        }
                                        local_a0 = plVar11;
                                        iVar4 = (int)*plVar10 + 1;
                                        if (iVar4 != 0) {
                                          *(int *)plVar10 = iVar4;
                                        }
                                        local_e8 = plVar10;
                                        if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) &&
                                           (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                          lVar16 = FUN_18000cc40(plVar7);
                                          (**(code **)(lVar16 + 0x30))(plVar7);
                                        }
                                        if (local_f0 == (longlong *)0x0) {
/* novels.services.state_reducer#33="hooks_by_id" */
                                          FUN_180370fe0(param_1,&local_res20,DAT_1803c55e8);
                                          FUN_18000d1d0(param_1,local_res20);
                                          uVar17 = 0x61;
                                          goto LAB_1802d916b;
                                        }
                                        plVar7 = (longlong *)
/* novels.services.state_reducer#3="get" */
                                                 FUN_18037f0c0(param_1,local_f0,DAT_1803c54f8);
                                        if (plVar7 == (longlong *)0x0) {
LAB_1802d87a7:
                                          FUN_18000ce30(param_1,&local_res20);
                                          uVar17 = 0x61;
                                          goto LAB_1802d916b;
                                        }
                                        plVar9 = (longlong *)FUN_1803749a0(plVar10);
                                        if (plVar9 == (longlong *)0x0) {
                                          FUN_18000ce30(param_1,&local_res20);
                                          if ((-1 < (int)*plVar7) &&
                                             (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                            lVar16 = FUN_18000cc40(plVar7);
                                            (**(code **)(lVar16 + 0x30))(plVar7);
                                          }
                                          uVar17 = 0x61;
                                          goto LAB_1802d916b;
                                        }
                                        *(undefined4 *)(plVar2 + 5) = 0x61;
                                        plVar11 = (longlong *)FUN_180376770(param_1,plVar7,plVar9);
                                        if ((-1 < (int)*plVar7) &&
                                           (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                          lVar16 = FUN_18000cc40(plVar7);
                                          (**(code **)(lVar16 + 0x30))(plVar7);
                                        }
                                        if ((-1 < (int)*plVar9) &&
                                           (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                          lVar16 = FUN_18000cc40(plVar9);
                                          (**(code **)(lVar16 + 0x30))(plVar9);
                                        }
                                        if (plVar11 == (longlong *)0x0) goto LAB_1802d87a7;
                                        local_f8 = plVar11;
                                        if (((plVar6 != (longlong *)0x0) && (-1 < (int)*plVar6)) &&
                                           (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                          lVar16 = FUN_18000cc40(plVar6);
                                          (**(code **)(lVar16 + 0x30))(plVar6);
                                        }
                                        iVar4 = FUN_180001510(plVar11);
                                        if (iVar4 == -1) {
                                          FUN_18000ce30(param_1,&local_res20);
                                          uVar17 = 0x62;
                                          goto LAB_1802d916b;
                                        }
                                        if (iVar4 != 0) {
                                          plVar6 = (longlong *)FUN_180370dc0(param_1,plVar11,0);
                                          if (plVar6 == (longlong *)0x0) {
                                            FUN_18000ce30(param_1,&local_res20);
                                            uVar17 = 100;
                                          }
                                          else {
                                            local_f8 = plVar6;
                                            if ((-1 < (int)*plVar11) &&
                                               (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                              lVar16 = FUN_18000cc40(plVar11);
                                              (**(code **)(lVar16 + 0x30))(plVar11);
                                            }
/* novels.services.state_reducer#46="status" ; novels.services.state_reducer#45="resolved" */
                                            PyDict_SetItem(plVar6,DAT_1803c5650,DAT_1803c5648);
/* novels.services.state_reducer#47="max" */
                                            uVar12 = FUN_18000d570(DAT_1803c5658);
                                            pcVar13 = (code *)FUN_18036f6e0(param_1,plVar6,
/* novels.services.state_reducer#48="lastAdvancedChapter" */
                                                                            DAT_1803c5660);
                                            if (pcVar13 == (code *)0x0) {
                                              pcVar13 = _Py_NoneStruct_exref;
                                            }
                                            iVar4 = FUN_180001510(pcVar13);
                                            if (iVar4 == -1) {
                                              FUN_18000ce30(param_1,&local_res20);
                                              if ((-1 < *(int *)pcVar13) &&
                                                 (*(longlong *)pcVar13 = *(longlong *)pcVar13 + -1,
                                                 *(longlong *)pcVar13 == 0)) {
                                                lVar16 = FUN_18000cc40(pcVar13);
                                                (**(code **)(lVar16 + 0x30))(pcVar13);
                                              }
                                              uVar17 = 0x67;
                                            }
                                            else {
                                              if (iVar4 != 1) {
                                                if ((-1 < *(int *)pcVar13) &&
                                                   (*(longlong *)pcVar13 = *(longlong *)pcVar13 + -1
                                                   , *(longlong *)pcVar13 == 0)) {
                                                  lVar16 = FUN_18000cc40(pcVar13);
                                                  (**(code **)(lVar16 + 0x30))(pcVar13);
                                                }
                                                pcVar13 = DAT_1803ab250;
                                                if (*(int *)DAT_1803ab250 + 1 != 0) {
                                                  *(int *)DAT_1803ab250 = *(int *)DAT_1803ab250 + 1;
                                                }
                                              }
                                              plVar7 = (longlong *)PyNumber_Long(pcVar13);
                                              if ((-1 < *(int *)pcVar13) &&
                                                 (*(longlong *)pcVar13 = *(longlong *)pcVar13 + -1,
                                                 *(longlong *)pcVar13 == 0)) {
                                                lVar16 = FUN_18000cc40(pcVar13);
                                                (**(code **)(lVar16 + 0x30))(pcVar13);
                                              }
                                              if (plVar7 == (longlong *)0x0) {
                                                FUN_18000ce30(param_1,&local_res20);
                                                uVar17 = 0x67;
                                              }
                                              else {
                                                *(undefined4 *)(plVar2 + 5) = 0x66;
                                                local_80 = local_98;
                                                local_88 = plVar7;
                                                plVar9 = (longlong *)
                                                         FUN_180377750(param_1,uVar12,&local_88);
                                                if ((-1 < (int)*plVar7) &&
                                                   (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                  lVar16 = FUN_18000cc40(plVar7);
                                                  (**(code **)(lVar16 + 0x30))(plVar7);
                                                }
                                                if (plVar9 == (longlong *)0x0) {
                                                  FUN_18000ce30(param_1,&local_res20);
                                                  uVar17 = 0x66;
                                                }
                                                else {
/* novels.services.state_reducer#48="lastAdvancedChapter" */
                                                  PyDict_SetItem(plVar6,DAT_1803c5660,plVar9);
                                                  if ((-1 < (int)*plVar9) &&
                                                     (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                    lVar16 = FUN_18000cc40(plVar9);
                                                    (**(code **)(lVar16 + 0x30))(plVar9);
                                                  }
                                                  plVar10 = (longlong *)FUN_1803749a0(plVar10);
                                                  if (plVar10 != (longlong *)0x0) {
                                                    cVar3 = FUN_18000d470(param_1,local_f0,plVar10,
                                                                          plVar6);
                                                    if ((-1 < (int)*plVar10) &&
                                                       (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                                      lVar16 = FUN_18000cc40(plVar10);
                                                      (**(code **)(lVar16 + 0x30))(plVar10);
                                                    }
                                                    if (cVar3 != '\0') {
                                                      cVar3 = FUN_18000d5b0();
                                                      if (cVar3 != '\0') goto LAB_1802d85fa;
                                                      FUN_18000ce30(param_1,&local_res20);
                                                      uVar17 = local_res18._0_4_;
                                                      goto LAB_1802d916b;
                                                    }
                                                  }
                                                  FUN_18000ce30(param_1,&local_res20);
                                                  uVar17 = 0x6a;
                                                }
                                              }
                                            }
                                          }
                                          goto LAB_1802d916b;
                                        }
LAB_1802d85fa:
                                        plVar10 = (longlong *)FUN_18000d430(plVar8);
                                      } while (plVar10 != (longlong *)0x0);
                                      uVar17 = local_res18._0_4_;
                                    }
                                    cVar3 = FUN_1800012a0(param_1);
                                    plVar10 = local_a0;
                                    if (cVar3 == '\0') {
                                      FUN_18000ce30(param_1,&local_res20);
LAB_1802d916b:
                                      lVar16 = local_res20;
                                      FUN_18000ce70(&local_res20);
                                      plVar10 = local_a0;
LAB_1802d910f:
                                      if (((plVar10 != (longlong *)0x0) && (-1 < (int)*plVar10)) &&
                                         (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                        lVar15 = FUN_18000cc40(plVar10);
                                        (**(code **)(lVar15 + 0x30))(plVar10);
                                      }
                                      plVar10 = local_100;
                                      if ((-1 < (int)*plVar8) &&
                                         (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                        lVar15 = FUN_18000cc40(plVar8);
                                        (**(code **)(lVar15 + 0x30))(plVar8);
                                        plVar10 = local_100;
                                      }
                                    }
                                    else {
                                      if (((local_a0 != (longlong *)0x0) && (-1 < (int)*local_a0))
                                         && (*local_a0 = *local_a0 + -1, *local_a0 == 0)) {
                                        lVar16 = FUN_18000cc40(local_a0);
                                        (**(code **)(lVar16 + 0x30))(plVar10);
                                      }
                                      if ((-1 < (int)*plVar8) &&
                                         (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                        lVar16 = FUN_18000cc40(plVar8);
                                        (**(code **)(lVar16 + 0x30))(plVar8);
                                      }
                                      plVar8 = (longlong *)
/* novels.services.state_reducer#3="get" */
                                               FUN_18037f0c0(param_1,local_d0,DAT_1803c54f8);
                                      if (plVar8 == (longlong *)0x0) {
                                        FUN_18000ce30(param_1,&local_res20);
                                        uVar17 = 0x6c;
                                        plVar10 = local_100;
                                        lVar16 = local_res20;
                                      }
                                      else {
                                        uVar17 = 0x6c;
                                        *(undefined4 *)(plVar2 + 5) = 0x6c;
                                        local_res18 = (longlong *)CONCAT44(local_res18._4_4_,0x6c);
                                        plVar6 = (longlong *)
                                                 FUN_1803770d0(param_1,plVar8,DAT_1803c5668);
                                        if ((-1 < (int)*plVar8) &&
                                           (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                          lVar16 = FUN_18000cc40(plVar8);
                                          (**(code **)(lVar16 + 0x30))(plVar8);
                                        }
                                        if (plVar6 == (longlong *)0x0) {
                                          FUN_18000ce30(param_1,&local_res20);
                                          plVar10 = local_100;
                                          lVar16 = local_res20;
                                        }
                                        else {
                                          iVar4 = FUN_180001510(plVar6);
                                          if (iVar4 == -1) {
                                            FUN_18000ce30(param_1,&local_res20);
                                            plVar10 = local_100;
                                            lVar16 = local_res20;
                                            if ((-1 < (int)*plVar6) &&
                                               (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                              lVar16 = FUN_18000cc40(plVar6);
                                              (**(code **)(lVar16 + 0x30))(plVar6);
                                              plVar10 = local_100;
                                              lVar16 = local_res20;
                                            }
                                          }
                                          else {
                                            if (iVar4 != 1) {
                                              if ((-1 < (int)*plVar6) &&
                                                 (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                lVar16 = FUN_18000cc40(plVar6);
                                                (**(code **)(lVar16 + 0x30))(plVar6);
                                              }
                                              plVar6 = (longlong *)FUN_180372160(param_1,0);
                                            }
                                            plVar8 = (longlong *)FUN_180001930(param_1,plVar6);
                                            if ((-1 < (int)*plVar6) &&
                                               (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                              lVar16 = FUN_18000cc40(plVar6);
                                              (**(code **)(lVar16 + 0x30))(plVar6);
                                            }
                                            if (plVar8 == (longlong *)0x0) {
                                              FUN_18000ce30(param_1,&local_res20);
                                              plVar10 = local_100;
                                              lVar16 = local_res20;
                                            }
                                            else {
                                              plVar10 = (longlong *)FUN_18000d430(plVar8);
                                              if (plVar10 != (longlong *)0x0) {
                                                do {
                                                  plVar9 = local_a8;
                                                  plVar7 = local_e8;
                                                  plVar6 = local_f8;
                                                  plVar11 = plVar10;
                                                  if (((local_a8 != (longlong *)0x0) &&
                                                      (-1 < (int)*local_a8)) &&
                                                     (*local_a8 = *local_a8 + -1, *local_a8 == 0)) {
                                                    local_a8 = plVar10;
                                                    lVar16 = FUN_18000cc40();
                                                    (**(code **)(lVar16 + 0x30))(plVar9);
                                                    plVar11 = local_a8;
                                                  }
                                                  local_a8 = plVar11;
                                                  iVar4 = (int)*plVar10 + 1;
                                                  if (iVar4 != 0) {
                                                    *(int *)plVar10 = iVar4;
                                                  }
                                                  local_e8 = plVar10;
                                                  if (((plVar7 != (longlong *)0x0) &&
                                                      (-1 < (int)*plVar7)) &&
                                                     (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                    lVar16 = FUN_18000cc40(plVar7);
                                                    (**(code **)(lVar16 + 0x30))(plVar7);
                                                  }
                                                  if (local_f0 == (longlong *)0x0) {
/* novels.services.state_reducer#33="hooks_by_id" */
                                                    FUN_180370fe0(param_1,&local_res20,DAT_1803c55e8
                                                                 );
                                                    FUN_18000d1d0(param_1,local_res20);
                                                    uVar17 = 0x6d;
                                                    goto LAB_1802d90f8;
                                                  }
                                                  plVar7 = (longlong *)
                                                           FUN_18037f0c0(param_1,local_f0,
/* novels.services.state_reducer#3="get" */
                                                                         DAT_1803c54f8);
                                                  if (plVar7 == (longlong *)0x0) {
LAB_1802d8dea:
                                                    FUN_18000ce30(param_1,&local_res20);
                                                    uVar17 = 0x6d;
                                                    goto LAB_1802d90f8;
                                                  }
                                                  plVar9 = (longlong *)FUN_1803749a0(plVar10);
                                                  if (plVar9 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_res20);
                                                    if ((-1 < (int)*plVar7) &&
                                                       (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                      lVar16 = FUN_18000cc40(plVar7);
                                                      (**(code **)(lVar16 + 0x30))(plVar7);
                                                    }
                                                    uVar17 = 0x6d;
                                                    goto LAB_1802d90f8;
                                                  }
                                                  *(undefined4 *)(plVar2 + 5) = 0x6d;
                                                  plVar11 = (longlong *)
                                                            FUN_180376770(param_1,plVar7,plVar9);
                                                  if ((-1 < (int)*plVar7) &&
                                                     (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                    lVar16 = FUN_18000cc40(plVar7);
                                                    (**(code **)(lVar16 + 0x30))(plVar7);
                                                  }
                                                  if ((-1 < (int)*plVar9) &&
                                                     (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                    lVar16 = FUN_18000cc40(plVar9);
                                                    (**(code **)(lVar16 + 0x30))(plVar9);
                                                  }
                                                  if (plVar11 == (longlong *)0x0)
                                                  goto LAB_1802d8dea;
                                                  local_f8 = plVar11;
                                                  if (((plVar6 != (longlong *)0x0) &&
                                                      (-1 < (int)*plVar6)) &&
                                                     (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                    lVar16 = FUN_18000cc40(plVar6);
                                                    (**(code **)(lVar16 + 0x30))(plVar6);
                                                  }
                                                  iVar4 = FUN_180001510(plVar11);
                                                  if (iVar4 == -1) {
                                                    FUN_18000ce30(param_1,&local_res20);
                                                    uVar17 = 0x6e;
                                                    goto LAB_1802d90f8;
                                                  }
                                                  if (iVar4 != 0) {
                                                    plVar6 = (longlong *)
                                                             FUN_180370dc0(param_1,plVar11,0);
                                                    if (plVar6 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_res20);
                                                      uVar17 = 0x70;
                                                    }
                                                    else {
                                                      local_f8 = plVar6;
                                                      if ((-1 < (int)*plVar11) &&
                                                         (*plVar11 = *plVar11 + -1, *plVar11 == 0))
                                                      {
                                                        lVar16 = FUN_18000cc40(plVar11);
                                                        (**(code **)(lVar16 + 0x30))(plVar11);
                                                      }
/* novels.services.state_reducer#46="status" */
                                                      PyDict_SetItem(plVar6,DAT_1803c5650,
/* novels.services.state_reducer#50="deferred" */
                                                                     DAT_1803c5670);
/* novels.services.state_reducer#47="max" */
                                                      uVar12 = FUN_18000d570(DAT_1803c5658);
                                                      pcVar13 = (code *)FUN_18036f6e0(param_1,plVar6
/* novels.services.state_reducer#48="lastAdvancedChapter" */
                                                                                      ,DAT_1803c5660
                                                                                     );
                                                      if (pcVar13 == (code *)0x0) {
                                                        pcVar13 = _Py_NoneStruct_exref;
                                                      }
                                                      iVar4 = FUN_180001510(pcVar13);
                                                      if (iVar4 == -1) {
                                                        FUN_18000ce30(param_1,&local_res20);
                                                        if ((-1 < *(int *)pcVar13) &&
                                                           (*(longlong *)pcVar13 =
                                                                 *(longlong *)pcVar13 + -1,
                                                           *(longlong *)pcVar13 == 0)) {
                                                          lVar16 = FUN_18000cc40(pcVar13);
                                                          (**(code **)(lVar16 + 0x30))(pcVar13);
                                                        }
                                                        uVar17 = 0x73;
                                                      }
                                                      else {
                                                        if (iVar4 != 1) {
                                                          if ((-1 < *(int *)pcVar13) &&
                                                             (*(longlong *)pcVar13 =
                                                                   *(longlong *)pcVar13 + -1,
                                                             *(longlong *)pcVar13 == 0)) {
                                                            lVar16 = FUN_18000cc40(pcVar13);
                                                            (**(code **)(lVar16 + 0x30))(pcVar13);
                                                          }
                                                          pcVar13 = DAT_1803ab250;
                                                          if (*(int *)DAT_1803ab250 + 1 != 0) {
                                                            *(int *)DAT_1803ab250 =
                                                                 *(int *)DAT_1803ab250 + 1;
                                                          }
                                                        }
                                                        plVar7 = (longlong *)PyNumber_Long(pcVar13);
                                                        if ((-1 < *(int *)pcVar13) &&
                                                           (*(longlong *)pcVar13 =
                                                                 *(longlong *)pcVar13 + -1,
                                                           *(longlong *)pcVar13 == 0)) {
                                                          lVar16 = FUN_18000cc40(pcVar13);
                                                          (**(code **)(lVar16 + 0x30))(pcVar13);
                                                        }
                                                        if (plVar7 == (longlong *)0x0) {
                                                          FUN_18000ce30(param_1,&local_res20);
                                                          uVar17 = 0x73;
                                                        }
                                                        else {
                                                          *(undefined4 *)(plVar2 + 5) = 0x72;
                                                          local_80 = local_98;
                                                          local_88 = plVar7;
                                                          plVar9 = (longlong *)
                                                                   FUN_180377750(param_1,uVar12,
                                                                                 &local_88);
                                                          if ((-1 < (int)*plVar7) &&
                                                             (*plVar7 = *plVar7 + -1, *plVar7 == 0))
                                                          {
                                                            lVar16 = FUN_18000cc40(plVar7);
                                                            (**(code **)(lVar16 + 0x30))(plVar7);
                                                          }
                                                          if (plVar9 == (longlong *)0x0) {
                                                            FUN_18000ce30(param_1,&local_res20);
                                                            uVar17 = 0x72;
                                                          }
                                                          else {
/* novels.services.state_reducer#48="lastAdvancedChapter" */
                                                            PyDict_SetItem(plVar6,DAT_1803c5660,
                                                                           plVar9);
                                                            if ((-1 < (int)*plVar9) &&
                                                               (*plVar9 = *plVar9 + -1, *plVar9 == 0
                                                               )) {
                                                              lVar16 = FUN_18000cc40(plVar9);
                                                              (**(code **)(lVar16 + 0x30))(plVar9);
                                                            }
                                                            plVar10 = (longlong *)
                                                                      FUN_1803749a0(plVar10);
                                                            if (plVar10 != (longlong *)0x0) {
                                                              cVar3 = FUN_18000d470(param_1,local_f0
                                                                                    ,plVar10,plVar6)
                                                              ;
                                                              if ((-1 < (int)*plVar10) &&
                                                                 (*plVar10 = *plVar10 + -1,
                                                                 *plVar10 == 0)) {
                                                                lVar16 = FUN_18000cc40(plVar10);
                                                                (**(code **)(lVar16 + 0x30))
                                                                          (plVar10);
                                                              }
                                                              if (cVar3 != '\0') {
                                                                cVar3 = FUN_18000d5b0();
                                                                if (cVar3 != '\0')
                                                                goto LAB_1802d8c2a;
                                                                FUN_18000ce30(param_1,&local_res20);
                                                                uVar17 = local_res18._0_4_;
                                                                goto LAB_1802d90f8;
                                                              }
                                                            }
                                                            FUN_18000ce30(param_1,&local_res20);
                                                            uVar17 = 0x76;
                                                          }
                                                        }
                                                      }
                                                    }
                                                    goto LAB_1802d90f8;
                                                  }
LAB_1802d8c2a:
                                                  plVar10 = (longlong *)FUN_18000d430(plVar8);
                                                } while (plVar10 != (longlong *)0x0);
                                                uVar17 = local_res18._0_4_;
                                              }
                                              cVar3 = FUN_1800012a0(param_1);
                                              plVar10 = local_a8;
                                              if (cVar3 == '\0') {
                                                FUN_18000ce30(param_1,&local_res20);
LAB_1802d90f8:
                                                lVar16 = local_res20;
                                                FUN_18000ce70(&local_res20);
                                                plVar10 = local_a8;
                                                goto LAB_1802d910f;
                                              }
                                              if (((local_a8 != (longlong *)0x0) &&
                                                  (-1 < (int)*local_a8)) &&
                                                 (*local_a8 = *local_a8 + -1, *local_a8 == 0)) {
                                                lVar16 = FUN_18000cc40(local_a8);
                                                (**(code **)(lVar16 + 0x30))(plVar10);
                                              }
                                              if ((-1 < (int)*plVar8) &&
                                                 (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                lVar16 = FUN_18000cc40(plVar8);
                                                (**(code **)(lVar16 + 0x30))(plVar8);
                                              }
/* novels.services.state_reducer#51="sorted" */
                                              uVar12 = FUN_18000d570(DAT_1803c5678);
                                              plVar8 = local_f0;
                                              if (local_f0 == (longlong *)0x0) {
/* novels.services.state_reducer#33="hooks_by_id" */
                                                FUN_180370fe0(param_1,&local_res20,DAT_1803c55e8);
                                                FUN_18000d1d0(param_1,local_res20);
                                                uVar17 = 0x79;
                                                plVar10 = local_100;
                                                lVar16 = local_res20;
                                              }
                                              else {
                                                plVar10 = (longlong *)
                                                          FUN_18037f0c0(param_1,local_f0,
/* novels.services.state_reducer#36="values" */
                                                                        DAT_1803c5600);
                                                if (plVar10 == (longlong *)0x0) {
                                                  FUN_18000ce30(param_1,&local_res20);
                                                  uVar17 = 0x79;
                                                  plVar10 = local_100;
                                                  lVar16 = local_res20;
                                                }
                                                else {
                                                  uVar17 = 0x79;
                                                  *(undefined4 *)(plVar2 + 5) = 0x79;
                                                  plVar6 = (longlong *)
                                                           FUN_180376210(param_1,plVar10);
                                                  if ((-1 < (int)*plVar10) &&
                                                     (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                                    lVar16 = FUN_18000cc40(plVar10);
                                                    (**(code **)(lVar16 + 0x30))(plVar10);
                                                  }
                                                  if (plVar6 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_res20);
                                                    plVar10 = local_100;
                                                    lVar16 = local_res20;
                                                  }
                                                  else {
                                                    plVar10 = (longlong *)FUN_1802e35f0(param_1);
                                                    uVar17 = 0x78;
                                                    *(undefined4 *)(plVar2 + 5) = 0x78;
                                                    local_res18 = plVar10;
                                                    local_108 = plVar6;
                                                    plVar7 = (longlong *)
                                                             FUN_18037d030(param_1,uVar12,&local_108
                                                                           ,&local_res18,
                                                                           DAT_1803c5690);
                                                    if ((-1 < (int)*plVar6) &&
                                                       (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                      lVar16 = FUN_18000cc40(plVar6);
                                                      (**(code **)(lVar16 + 0x30))(plVar6);
                                                    }
                                                    if ((-1 < (int)*plVar10) &&
                                                       (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                                      lVar16 = FUN_18000cc40(plVar10);
                                                      (**(code **)(lVar16 + 0x30))(plVar10);
                                                    }
                                                    if (plVar7 != (longlong *)0x0) {
                                                      FUN_180001d80(param_1);
/* novels.services.state_reducer#7="hooks" */
                                                      uVar12 = DAT_1803c5518;
                                                      uVar14 = _PyDict_NewPresized(1);
                                                      PyDict_SetItem(uVar14,uVar12,plVar7);
                                                      if ((-1 < (int)*plVar8) &&
                                                         (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                        lVar16 = FUN_18000cc40(plVar8);
                                                        (**(code **)(lVar16 + 0x30))(plVar8);
                                                      }
                                                      plVar8 = local_d0;
                                                      if ((-1 < (int)*local_d0) &&
                                                         (*local_d0 = *local_d0 + -1, *local_d0 == 0
                                                         )) {
                                                        lVar16 = FUN_18000cc40(local_d0);
                                                        (**(code **)(lVar16 + 0x30))(plVar8);
                                                      }
                                                      plVar8 = local_100;
                                                      if (((local_100 != (longlong *)0x0) &&
                                                          (-1 < (int)*local_100)) &&
                                                         (*local_100 = *local_100 + -1,
                                                         *local_100 == 0)) {
                                                        lVar16 = FUN_18000cc40(local_100);
                                                        (**(code **)(lVar16 + 0x30))(plVar8);
                                                      }
                                                      plVar8 = local_e8;
                                                      if (((local_e8 != (longlong *)0x0) &&
                                                          (-1 < (int)*local_e8)) &&
                                                         (*local_e8 = *local_e8 + -1, *local_e8 == 0
                                                         )) {
                                                        lVar16 = FUN_18000cc40(local_e8);
                                                        (**(code **)(lVar16 + 0x30))(plVar8);
                                                      }
                                                      plVar8 = local_f8;
                                                      if (((local_f8 != (longlong *)0x0) &&
                                                          (-1 < (int)*local_f8)) &&
                                                         (*local_f8 = *local_f8 + -1, *local_f8 == 0
                                                         )) {
                                                        lVar16 = FUN_18000cc40(local_f8);
                                                        (**(code **)(lVar16 + 0x30))(plVar8);
                                                      }
                                                      plVar8 = local_c8;
                                                      if (((local_c8 != (longlong *)0x0) &&
                                                          (-1 < (int)*local_c8)) &&
                                                         (*local_c8 = *local_c8 + -1, *local_c8 == 0
                                                         )) {
                                                        lVar16 = FUN_18000cc40(local_c8);
                                                        (**(code **)(lVar16 + 0x30))(plVar8);
                                                      }
                                                      plVar8 = local_c0;
                                                      if (((local_c0 != (longlong *)0x0) &&
                                                          (-1 < (int)*local_c0)) &&
                                                         (*local_c0 = *local_c0 + -1, *local_c0 == 0
                                                         )) {
                                                        lVar16 = FUN_18000cc40(local_c0);
                                                        (**(code **)(lVar16 + 0x30))(plVar8);
                                                      }
                                                      if ((-1 < (int)*plVar7) &&
                                                         (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                        lVar16 = FUN_18000cc40(plVar7);
                                                        (**(code **)(lVar16 + 0x30))(plVar7);
                                                      }
                                                      plVar8 = local_90;
                                                      if ((-1 < (int)*local_90) &&
                                                         (*local_90 = *local_90 + -1, *local_90 == 0
                                                         )) {
                                                        lVar16 = FUN_18000cc40(local_90);
                                                        (**(code **)(lVar16 + 0x30))(plVar8);
                                                      }
                                                      plVar8 = local_68;
                                                      if ((-1 < (int)*local_68) &&
                                                         (*local_68 = *local_68 + -1, *local_68 == 0
                                                         )) {
                                                        lVar16 = FUN_18000cc40(local_68);
                                                        (**(code **)(lVar16 + 0x30))(plVar8);
                                                      }
                                                      plVar8 = local_98;
                                                      if ((int)*local_98 < 0) {
                                                        return uVar14;
                                                      }
                                                      *local_98 = *local_98 + -1;
                                                      if (*local_98 != 0) {
                                                        return uVar14;
                                                      }
                                                      lVar16 = FUN_18000cc40(local_98);
                                                      (**(code **)(lVar16 + 0x30))(plVar8);
                                                      return uVar14;
                                                    }
                                                    FUN_18000ce30(param_1,&local_res20);
                                                    plVar10 = local_100;
                                                    lVar16 = local_res20;
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
                    }
                  }
                  goto LAB_1802d929c;
                }
LAB_1802d91f4:
                local_res20 = *(longlong *)(param_1 + 0x60);
                *(undefined8 *)(param_1 + 0x60) = 0;
LAB_1802d9207:
                lVar16 = local_res20;
                if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                  (**(code **)(plVar8[1] + 0x30))(plVar8);
                }
                if ((-1 < (int)*local_f0) && (*local_f0 = *local_f0 + -1, *local_f0 == 0)) {
                  (**(code **)(local_f0[1] + 0x30))();
                }
                if (plVar10 != (longlong *)0x0) {
                  if ((int)*plVar10 < 0) {
                    if (plVar10 == (longlong *)0x0) goto LAB_1802d9286;
                  }
                  else {
                    *plVar10 = *plVar10 + -1;
                    if (*plVar10 == 0) {
                      (**(code **)(plVar10[1] + 0x30))(plVar10);
                    }
                  }
                  if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                    (**(code **)(plVar10[1] + 0x30))(plVar10);
                  }
                }
              }
            }
          }
        }
LAB_1802d9286:
        uVar17 = 0x44;
        plVar10 = local_100;
      }
      else {
/* novels.services.state_reducer#3="get" */
        plVar8 = (longlong *)FUN_18037f0c0(param_1,plVar8,DAT_1803c54f8);
        if (plVar8 == (longlong *)0x0) {
          lVar16 = *(longlong *)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          plVar10 = (longlong *)0x0;
        }
        else {
          *(undefined4 *)(plVar2 + 5) = 0x43;
          local_d0 = (longlong *)FUN_1803770d0(param_1,plVar8,DAT_1803c55c8);
          if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
            (**(code **)(plVar8[1] + 0x30))(plVar8);
          }
          if (local_d0 != (longlong *)0x0) goto LAB_1802d6d2f;
          lVar16 = *(longlong *)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          plVar10 = (longlong *)0x0;
        }
      }
    }
  }
LAB_1802d929c:
  local_res20 = lVar16;
  piVar1 = *(int **)(local_res20 + 0x28);
  if (piVar1 == (int *)0x0) {
    lVar16 = FUN_18038e2a0(plVar2,uVar17);
  }
  else {
    if (*(longlong **)(piVar1 + 6) == plVar2) goto LAB_1802d92e6;
    lVar16 = FUN_18038e2a0(plVar2,uVar17);
    *(int **)(lVar16 + 0x10) = piVar1;
    if (*piVar1 + 1 != 0) {
      *piVar1 = *piVar1 + 1;
    }
  }
  FUN_180001300(&local_res20,lVar16);
LAB_1802d92e6:
  plVar9 = local_98;
  plVar7 = local_e8;
  plVar6 = local_f8;
  FUN_18036d0a0(plVar2,"ooooooooooo",local_90,local_68,local_98,local_78,local_70,plVar10,local_e8,
                local_f8,local_c8,local_c0,0);
  plVar8 = DAT_1803c5480;
  if (plVar2 == DAT_1803c5480) {
    if ((-1 < (int)*DAT_1803c5480) && (*DAT_1803c5480 = *DAT_1803c5480 + -1, *plVar8 == 0)) {
      (**(code **)(plVar8[1] + 0x30))();
    }
    DAT_1803c5480 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  lVar16 = local_res20;
  local_res20 = 0;
  if (((local_78 != (longlong *)0x0) && (-1 < (int)*local_78)) &&
     (*local_78 = *local_78 + -1, *local_78 == 0)) {
    (**(code **)(local_78[1] + 0x30))();
  }
  if (((local_70 != (longlong *)0x0) && (-1 < (int)*local_70)) &&
     (*local_70 = *local_70 + -1, *local_70 == 0)) {
    (**(code **)(local_70[1] + 0x30))();
  }
  if (((plVar10 != (longlong *)0x0) && (-1 < (int)*plVar10)) &&
     (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
    (**(code **)(plVar10[1] + 0x30))(plVar10);
  }
  if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)
     ) {
    (**(code **)(plVar7[1] + 0x30))(plVar7);
  }
  if (((plVar6 != (longlong *)0x0) && (-1 < (int)*plVar6)) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)
     ) {
    (**(code **)(plVar6[1] + 0x30))(plVar6);
  }
  if (((local_c8 != (longlong *)0x0) && (-1 < (int)*local_c8)) &&
     (*local_c8 = *local_c8 + -1, *local_c8 == 0)) {
    (**(code **)(local_c8[1] + 0x30))();
  }
  if (((local_c0 != (longlong *)0x0) && (-1 < (int)*local_c0)) &&
     (*local_c0 = *local_c0 + -1, *local_c0 == 0)) {
    (**(code **)(local_c0[1] + 0x30))();
  }
  local_res20 = lVar16;
  if ((-1 < (int)*local_90) && (*local_90 = *local_90 + -1, *local_90 == 0)) {
    (**(code **)(local_90[1] + 0x30))();
  }
  if ((-1 < (int)*local_68) && (*local_68 = *local_68 + -1, *local_68 == 0)) {
    (**(code **)(local_68[1] + 0x30))();
  }
  if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
    (**(code **)(plVar9[1] + 0x30))(plVar9);
  }
  FUN_18000ce40(param_1,&local_res20);
  return 0;
}
