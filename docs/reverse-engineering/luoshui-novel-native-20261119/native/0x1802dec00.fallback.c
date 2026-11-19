/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x1802dec00 */

longlong * FUN_1802dec00(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  longlong *plVar2;
  longlong *plVar3;
  longlong lVar4;
  int *piVar5;
  longlong *plVar6;
  undefined8 uVar7;
  char cVar8;
  int iVar9;
  undefined4 uVar10;
  longlong *plVar11;
  longlong *plVar12;
  longlong *plVar13;
  longlong *plVar14;
  longlong *plVar15;
  longlong *plVar16;
  longlong lVar17;
  undefined8 uVar18;
  longlong *plVar19;
  longlong *plVar20;
  code *pcVar21;
  longlong *plVar22;
  code *pcVar23;
  code *pcVar24;
  longlong *plVar25;
  ulonglong uVar26;
  longlong *local_res18;
  undefined4 local_res20 [2];
  longlong *local_e0;
  longlong *local_d0;
  code *local_c8;
  longlong *local_b8;
  longlong *local_b0;
  longlong *local_a8;
  longlong *local_98;
  longlong *local_88;
  longlong *local_80;
  undefined8 local_50;
  longlong *local_48;
  
  plVar1 = (longlong *)*param_3;
  plVar15 = (longlong *)0x0;
  plVar2 = (longlong *)param_3[2];
  plVar25 = (longlong *)0x0;
  plVar3 = (longlong *)param_3[1];
  local_98 = (longlong *)0x0;
  local_a8 = (longlong *)0x0;
  local_88 = (longlong *)0x0;
  local_80 = (longlong *)0x0;
  local_d0 = (longlong *)0x0;
  local_b8 = (longlong *)0x0;
  local_c8 = (code *)0x0;
  local_e0 = (longlong *)0x0;
  plVar11 = (longlong *)FUN_18036d6d0();
  plVar6 = DAT_1803c53c0;
  local_b0 = (longlong *)0x0;
  if (DAT_1803c53c0 == (longlong *)0x0) {
LAB_1802decad:
    DAT_1803c53c0 = (longlong *)FUN_18036cd10(param_1,DAT_1803c5498,DAT_1803ca450,0x60);
  }
  else if ((1 < *DAT_1803c53c0) || (DAT_1803c53c0[2] != 0)) {
    if ((DAT_1803c53c0 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803c53c0 && (*DAT_1803c53c0 = *DAT_1803c53c0 + -1, *plVar6 == 0)))) {
      (**(code **)(plVar6[1] + 0x30))(plVar6);
    }
    goto LAB_1802decad;
  }
  plVar6 = DAT_1803c53c0;
  FUN_180001d00(param_1,DAT_1803c53c0);
  plVar12 = (longlong *)FUN_18037f0c0();
  if (plVar12 == (longlong *)0x0) {
    plVar12 = *(longlong **)(param_1 + 0x60);
    uVar26 = 0xf4;
    *(undefined8 *)(param_1 + 0x60) = 0;
    local_e0 = plVar15;
  }
  else {
    *(undefined4 *)(plVar6 + 5) = 0xf4;
    plVar13 = (longlong *)FUN_1803770d0(param_1,plVar12,DAT_1803c5788);
    if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
      (**(code **)(plVar12[1] + 0x30))(plVar12);
    }
    if (plVar13 == (longlong *)0x0) {
      plVar12 = *(longlong **)(param_1 + 0x60);
      uVar26 = 0xf4;
      *(undefined8 *)(param_1 + 0x60) = 0;
      local_e0 = plVar25;
    }
    else {
      plVar12 = (longlong *)FUN_18037f0c0();
      local_98 = plVar13;
      if (plVar12 == (longlong *)0x0) {
        plVar12 = *(longlong **)(param_1 + 0x60);
        uVar26 = 0xf5;
        *(undefined8 *)(param_1 + 0x60) = 0;
        local_e0 = plVar25;
      }
      else {
        *(undefined4 *)(plVar6 + 5) = 0xf5;
        plVar14 = (longlong *)FUN_1803770d0(param_1,plVar12,DAT_1803c5500);
        if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
          (**(code **)(plVar12[1] + 0x30))(plVar12);
        }
        if (plVar14 == (longlong *)0x0) {
          plVar12 = *(longlong **)(param_1 + 0x60);
          uVar26 = 0xf5;
          *(undefined8 *)(param_1 + 0x60) = 0;
          local_e0 = plVar15;
        }
        else {
          iVar9 = FUN_180001510(plVar14);
          if (iVar9 == -1) {
            plVar12 = *(longlong **)(param_1 + 0x60);
            *(undefined8 *)(param_1 + 0x60) = 0;
            if ((-1 < (int)*plVar14) && (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
              (**(code **)(plVar14[1] + 0x30))(plVar14);
            }
            uVar26 = 0xf5;
            local_e0 = plVar25;
          }
          else if (iVar9 == 1) {
LAB_1802def13:
            plVar25 = (longlong *)PyNumber_Long(plVar14);
            if ((-1 < (int)*plVar14) && (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
              (**(code **)(plVar14[1] + 0x30))(plVar14);
            }
            if (plVar25 == (longlong *)0x0) {
              plVar12 = *(longlong **)(param_1 + 0x60);
              uVar26 = 0xf5;
              *(undefined8 *)(param_1 + 0x60) = 0;
              local_e0 = (longlong *)0x0;
            }
            else {
              iVar9 = PyObject_IsInstance(plVar13,PyDict_Type_exref);
              local_a8 = plVar25;
              if (iVar9 != -1) {
                pcVar23 = _Py_FalseStruct_exref;
                if (iVar9 != 0) {
                  pcVar23 = _Py_TrueStruct_exref;
                }
                iVar9 = FUN_180001510(pcVar23);
                uVar18 = DAT_1803c54e8;
                if (iVar9 != -1) {
                  if (iVar9 != 0) {
                    plVar12 = (longlong *)FUN_18037f0c0();
                    local_res18 = (longlong *)CONCAT44(local_res18._4_4_,0xf9);
                    if (plVar12 == (longlong *)0x0) {
                      plVar12 = *(longlong **)(param_1 + 0x60);
                      *(undefined8 *)(param_1 + 0x60) = 0;
                      uVar26 = 0xf9;
                    }
                    else {
                      *(undefined4 *)(plVar6 + 5) = 0xf9;
                      plVar14 = (longlong *)FUN_1803770d0(param_1,plVar12,DAT_1803c5510);
                      if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                        (**(code **)(plVar12[1] + 0x30))(plVar12);
                      }
                      if (plVar14 == (longlong *)0x0) {
                        plVar12 = *(longlong **)(param_1 + 0x60);
                        *(undefined8 *)(param_1 + 0x60) = 0;
                        uVar26 = (ulonglong)local_res18 & 0xffffffff;
                      }
                      else {
                        iVar9 = FUN_180001510(plVar14);
                        if (iVar9 == -1) {
                          plVar12 = *(longlong **)(param_1 + 0x60);
                          *(undefined8 *)(param_1 + 0x60) = 0;
                          if (((int)*plVar14 < 0) || (*plVar14 = *plVar14 + -1, *plVar14 != 0)) {
LAB_1802df4ac:
                            uVar26 = (ulonglong)local_res18 & 0xffffffff;
                          }
                          else {
                            (**(code **)(plVar14[1] + 0x30))(plVar14);
                            uVar26 = (ulonglong)local_res18 & 0xffffffff;
                          }
                        }
                        else {
                          if (iVar9 != 1) {
                            if ((-1 < (int)*plVar14) && (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                              (**(code **)(plVar14[1] + 0x30))(plVar14);
                            }
                            plVar14 = (longlong *)FUN_180372160(param_1,0);
                          }
                          plVar16 = (longlong *)FUN_180001930();
                          if ((-1 < (int)*plVar14) && (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                            (**(code **)(plVar14[1] + 0x30))(plVar14);
                          }
                          if (plVar16 != (longlong *)0x0) {
                            plVar14 = (longlong *)FUN_180372160(param_1,0);
                            do {
                              pcVar23 = *(code **)(plVar16[1] + 0xe0);
                              if (pcVar23 == (code *)0x0) {
                                PyErr_Format(*(undefined8 *)PyExc_TypeError_exref,
                                             "\'%s\' object is not an iterator",
                                             *(undefined8 *)(plVar16[1] + 0x18));
LAB_1802df3db:
                                plVar12 = *(longlong **)(param_1 + 0x60);
                                if (plVar12 != (longlong *)0x0) {
                                  plVar19 = (longlong *)plVar12[1];
                                  if ((*(uint *)((longlong *)plVar12[1] + 0x15) & 0x40000000) == 0)
                                  {
                                    plVar19 = plVar12;
                                  }
                                  if ((plVar19 != *(longlong **)PyExc_StopIteration_exref) &&
                                     (((-1 < *(int *)(plVar19[1] + 0xa8) ||
                                       ((*(uint *)(plVar19 + 0x15) & 0x40000000) == 0)) ||
                                      (cVar8 = FUN_180382640(), cVar8 != '\x01'))))
                                  goto LAB_1802df433;
                                  *(undefined8 *)(param_1 + 0x60) = 0;
                                  if ((-1 < (int)*plVar12) &&
                                     (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                    (**(code **)(plVar12[1] + 0x30))(plVar12);
                                  }
                                }
                                iVar9 = (int)*plVar14 + 1;
                                if (iVar9 != 0) {
                                  *(int *)plVar14 = iVar9;
                                }
                                if ((-1 < (int)*plVar16) &&
                                   (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                  (**(code **)(plVar16[1] + 0x30))(plVar16);
                                }
                                if ((-1 < (int)*plVar14) &&
                                   (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                                  (**(code **)(plVar14[1] + 0x30))(plVar14);
                                }
                                if (((plVar15 != (longlong *)0x0) && (-1 < (int)*plVar15)) &&
                                   (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                                  (**(code **)(plVar15[1] + 0x30))(plVar15);
                                }
                                if (((plVar15 != (longlong *)0x0) && (-1 < (int)*plVar15)) &&
                                   (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                                  (**(code **)(plVar15[1] + 0x30))(plVar15);
                                }
                                iVar9 = FUN_1803836d0(plVar3,DAT_1803c5790);
                                local_88 = plVar14;
                                if (iVar9 == -1) {
                                  uVar26 = 0x103;
                                  goto LAB_1802e0893;
                                }
                                uVar18 = DAT_1803c57a0;
                                if (iVar9 == 1) {
                                  uVar18 = DAT_1803c5798;
                                }
                                local_80 = (longlong *)FUN_18036ff10(param_1,uVar18);
                                plVar15 = (longlong *)FUN_18036fa60();
                                plVar16 = (longlong *)FUN_180001930();
                                if ((-1 < (int)*plVar15) &&
                                   (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                                  (**(code **)(plVar15[1] + 0x30))(plVar15);
                                }
                                if (plVar16 != (longlong *)0x0) {
                                  local_res18 = (longlong *)CONCAT44(local_res18._4_4_,0x113);
                                  goto LAB_1802df5e0;
                                }
                                uVar26 = 0x10e;
                                goto LAB_1802e0893;
                              }
                              plVar12 = (longlong *)(*pcVar23)(plVar16);
                              if (plVar12 == (longlong *)0x0) goto LAB_1802df3db;
                              if (((plVar15 != (longlong *)0x0) && (-1 < (int)*plVar15)) &&
                                 (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                                (**(code **)(plVar15[1] + 0x30))();
                              }
                              iVar9 = (int)*plVar12 + 1;
                              if (iVar9 != 0) {
                                *(int *)plVar12 = iVar9;
                              }
                              if (((plVar15 != (longlong *)0x0) && (-1 < (int)*plVar15)) &&
                                 (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                                (**(code **)(plVar15[1] + 0x30))();
                              }
                              iVar9 = PyObject_IsInstance(plVar12,PyDict_Type_exref);
                              plVar15 = plVar12;
                              if (iVar9 == -1) goto LAB_1802df3ba;
                              if (iVar9 != 0) {
                                lVar17 = FUN_180370dc0(param_1,plVar12);
                                if (lVar17 == 0) goto LAB_1802df3ba;
                                lVar4 = plVar14[2];
                                cVar8 = FUN_1803722b0(plVar14,lVar4 + 1);
                                if (cVar8 == '\0') goto LAB_1802df3ba;
                                *(longlong *)(plVar14[3] + lVar4 * 8) = lVar17;
                              }
                              lVar17 = *(longlong *)(param_1 + 0x10);
                              if (((*(int *)(*(longlong *)(lVar17 + 0x60) + 0x4d4) != 0) ||
                                  (*(int *)(lVar17 + 0x1a0) != 0)) &&
                                 ((iVar9 = Py_MakePendingCalls(), iVar9 < 0 &&
                                  (*(longlong *)(param_1 + 0x60) != 0)))) goto LAB_1802df3ba;
                              if (*(int *)(lVar17 + 0x174) != 0) {
                                PyEval_SaveThread();
                                PyEval_AcquireThread(param_1);
                              }
                            } while (*(longlong *)(param_1 + 0x80) == 0);
                            plVar25 = *(longlong **)(param_1 + 0x60);
                            *(undefined8 *)(param_1 + 0x80) = 0;
                            uVar18 = FUN_180373490(param_1);
                            *(undefined8 *)(param_1 + 0x60) = uVar18;
                            if (((plVar25 != (longlong *)0x0) && (-1 < (int)*plVar25)) &&
                               (*plVar25 = *plVar25 + -1, *plVar25 == 0)) {
                              (**(code **)(plVar25[1] + 0x30))(plVar25);
                            }
LAB_1802df3ba:
                            plVar12 = *(longlong **)(param_1 + 0x60);
LAB_1802df433:
                            *(undefined8 *)(param_1 + 0x60) = 0;
                            if ((-1 < (int)*plVar16) && (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                              (**(code **)(plVar16[1] + 0x30))(plVar16);
                            }
                            if ((-1 < (int)*plVar14) && (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                              (**(code **)(plVar14[1] + 0x30))();
                            }
                            if (((plVar15 != (longlong *)0x0) && (-1 < (int)*plVar15)) &&
                               (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                              (**(code **)(plVar15[1] + 0x30))(plVar15);
                            }
                            if (((plVar15 != (longlong *)0x0) && (-1 < (int)*plVar15)) &&
                               (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                              (**(code **)(plVar15[1] + 0x30))(plVar15);
                            }
                            goto LAB_1802df4ac;
                          }
                          plVar12 = *(longlong **)(param_1 + 0x60);
                          *(undefined8 *)(param_1 + 0x60) = 0;
                          uVar26 = (ulonglong)local_res18 & 0xffffffff;
                        }
                      }
                    }
                    goto LAB_1802e08a9;
                  }
                  plVar15 = (longlong *)_PyDict_NewPresized(2);
                  PyDict_SetItem(plVar15,uVar18,plVar25);
                  uVar18 = DAT_1803c5508;
                  plVar12 = (longlong *)FUN_18037f0c0();
                  if (plVar12 == (longlong *)0x0) {
                    plVar12 = *(longlong **)(param_1 + 0x60);
                    *(undefined8 *)(param_1 + 0x60) = 0;
                  }
                  else {
                    *(undefined4 *)(plVar6 + 5) = 0xf7;
                    plVar14 = (longlong *)FUN_1803770d0(param_1,plVar12,DAT_1803c5510);
                    if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                      (**(code **)(plVar12[1] + 0x30))(plVar12);
                    }
                    if (plVar14 != (longlong *)0x0) {
                      iVar9 = FUN_180001510(plVar14);
                      if (iVar9 == -1) {
                        plVar12 = *(longlong **)(param_1 + 0x60);
                        *(undefined8 *)(param_1 + 0x60) = 0;
                        if ((-1 < (int)*plVar14) && (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                          (**(code **)(plVar14[1] + 0x30))(plVar14);
                        }
                        goto LAB_1802df0dc;
                      }
                      if (iVar9 != 1) {
                        if ((-1 < (int)*plVar14) && (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                          (**(code **)(plVar14[1] + 0x30))(plVar14);
                        }
                        plVar14 = (longlong *)FUN_180372160(param_1,0);
                      }
                      plVar12 = (longlong *)FUN_180372690(param_1,plVar14);
                      if ((-1 < (int)*plVar14) && (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                        (**(code **)(plVar14[1] + 0x30))(plVar14);
                      }
                      if (plVar12 != (longlong *)0x0) {
                        PyDict_SetItem(plVar15,uVar18,plVar12);
                        if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                          (**(code **)(plVar12[1] + 0x30))(plVar12);
                        }
                        FUN_180001d80(param_1);
LAB_1802e0bd3:
                        if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                          (**(code **)(plVar13[1] + 0x30))();
                        }
                        if ((-1 < (int)*plVar25) && (*plVar25 = *plVar25 + -1, *plVar25 == 0)) {
                          (**(code **)(plVar25[1] + 0x30))(plVar25);
                        }
                        if (((local_88 != (longlong *)0x0) && (-1 < (int)*local_88)) &&
                           (*local_88 = *local_88 + -1, *local_88 == 0)) {
                          (**(code **)(local_88[1] + 0x30))();
                        }
                        if (((local_80 != (longlong *)0x0) && (-1 < (int)*local_80)) &&
                           (*local_80 = *local_80 + -1, *local_80 == 0)) {
                          (**(code **)(local_80[1] + 0x30))();
                        }
                        if (((local_d0 != (longlong *)0x0) && (-1 < (int)*local_d0)) &&
                           (*local_d0 = *local_d0 + -1, *local_d0 == 0)) {
                          (**(code **)(local_d0[1] + 0x30))();
                        }
                        if (((local_b8 != (longlong *)0x0) && (-1 < (int)*local_b8)) &&
                           (*local_b8 = *local_b8 + -1, *local_b8 == 0)) {
                          (**(code **)(local_b8[1] + 0x30))();
                        }
                        if (((local_c8 != (code *)0x0) && (-1 < *(int *)local_c8)) &&
                           (*(longlong *)local_c8 = *(longlong *)local_c8 + -1,
                           *(longlong *)local_c8 == 0)) {
                          (**(code **)(*(longlong *)(local_c8 + 8) + 0x30))();
                        }
                        if (((local_e0 != (longlong *)0x0) && (-1 < (int)*local_e0)) &&
                           (*local_e0 = *local_e0 + -1, *local_e0 == 0)) {
                          (**(code **)(local_e0[1] + 0x30))();
                        }
                        if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                          (**(code **)(plVar11[1] + 0x30))();
                        }
                        if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
                          (**(code **)(plVar1[1] + 0x30))(plVar1);
                        }
                        if ((-1 < (int)*plVar3) && (*plVar3 = *plVar3 + -1, *plVar3 == 0)) {
                          (**(code **)(plVar3[1] + 0x30))();
                        }
                        if ((int)*plVar2 < 0) {
                          return plVar15;
                        }
                        *plVar2 = *plVar2 + -1;
                        if (*plVar2 != 0) {
                          return plVar15;
                        }
                        (**(code **)(plVar2[1] + 0x30))(plVar2);
                        return plVar15;
                      }
                    }
                    plVar12 = *(longlong **)(param_1 + 0x60);
                    *(undefined8 *)(param_1 + 0x60) = 0;
                  }
LAB_1802df0dc:
                  uVar26 = 0xf7;
                  if ((-1 < (int)*plVar15) && (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                    (**(code **)(plVar15[1] + 0x30))(plVar15);
                  }
                  goto LAB_1802e08a9;
                }
              }
              plVar12 = *(longlong **)(param_1 + 0x60);
              uVar26 = 0xf6;
              *(undefined8 *)(param_1 + 0x60) = 0;
              local_e0 = (longlong *)0x0;
            }
          }
          else {
            if ((-1 < (int)*plVar14) && (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
              (**(code **)(plVar14[1] + 0x30))(plVar14);
            }
            plVar12 = (longlong *)FUN_18037f0c0();
            if (plVar12 == (longlong *)0x0) {
              plVar12 = *(longlong **)(param_1 + 0x60);
              uVar26 = 0xf5;
              *(undefined8 *)(param_1 + 0x60) = 0;
              local_e0 = plVar25;
            }
            else {
              *(undefined4 *)(plVar6 + 5) = 0xf5;
              plVar14 = (longlong *)FUN_1803770d0(param_1,plVar12,DAT_1803c5500);
              if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                (**(code **)(plVar12[1] + 0x30))(plVar12);
              }
              if (plVar14 == (longlong *)0x0) {
                plVar12 = *(longlong **)(param_1 + 0x60);
                *(undefined8 *)(param_1 + 0x60) = 0;
                uVar26 = 0xf5;
                local_e0 = (longlong *)0x0;
              }
              else {
                iVar9 = FUN_180001510(plVar14);
                if (iVar9 != -1) {
                  if (iVar9 != 1) {
                    if ((-1 < (int)*plVar14) && (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                      (**(code **)(plVar14[1] + 0x30))(plVar14);
                    }
                    plVar14 = DAT_1803ab250;
                    iVar9 = (int)*DAT_1803ab250 + 1;
                    if (iVar9 != 0) {
                      *(int *)DAT_1803ab250 = iVar9;
                    }
                  }
                  goto LAB_1802def13;
                }
                plVar12 = *(longlong **)(param_1 + 0x60);
                *(undefined8 *)(param_1 + 0x60) = 0;
                if ((-1 < (int)*plVar14) && (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                  (**(code **)(plVar14[1] + 0x30))(plVar14);
                }
                uVar26 = 0xf5;
                local_e0 = (longlong *)0x0;
              }
            }
          }
        }
      }
    }
  }
  goto LAB_1802e08a9;
LAB_1802e0893:
  plVar12 = *(longlong **)(param_1 + 0x60);
  *(undefined8 *)(param_1 + 0x60) = 0;
  goto LAB_1802e08a9;
LAB_1802df5e0:
  do {
    do {
      pcVar23 = *(code **)(plVar16[1] + 0xe0);
      if (pcVar23 == (code *)0x0) {
        PyErr_Format(*(undefined8 *)PyExc_TypeError_exref,"\'%s\' object is not an iterator",
                     *(undefined8 *)(plVar16[1] + 0x18));
LAB_1802e065c:
        plVar12 = *(longlong **)(param_1 + 0x60);
        if (plVar12 != (longlong *)0x0) {
          plVar15 = (longlong *)plVar12[1];
          if ((*(uint *)((longlong *)plVar12[1] + 0x15) & 0x40000000) == 0) {
            plVar15 = plVar12;
          }
          if ((plVar15 != *(longlong **)PyExc_StopIteration_exref) &&
             (((-1 < *(int *)(plVar15[1] + 0xa8) || ((*(uint *)(plVar15 + 0x15) & 0x40000000) == 0))
              || (cVar8 = FUN_180382640(), cVar8 != '\x01')))) {
            *(undefined8 *)(param_1 + 0x60) = 0;
            uVar26 = 0x10e;
            if (local_b0 == (longlong *)0x0) goto LAB_1802e021f;
            goto LAB_1802e0203;
          }
          *(undefined8 *)(param_1 + 0x60) = 0;
          if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
            (**(code **)(plVar12[1] + 0x30))(plVar12);
          }
        }
        if (((local_b0 != (longlong *)0x0) && (-1 < (int)*local_b0)) &&
           (*local_b0 = *local_b0 + -1, *local_b0 == 0)) {
          (**(code **)(local_b0[1] + 0x30))(local_b0);
        }
        if ((-1 < (int)*plVar16) && (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
          (**(code **)(plVar16[1] + 0x30))(plVar16);
        }
        if (plVar14 == (longlong *)0x0) {
          plVar15 = (longlong *)
                    PyUnicode_FromFormat
                              ("cannot access local variable \'%U\' where it is not associated with a value"
                               ,DAT_1803c57b0);
          plVar12 = (longlong *)
                    FUN_180376770(param_1,*(undefined8 *)PyExc_UnboundLocalError_exref,plVar15);
          if ((-1 < (int)*plVar15) && (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
            (**(code **)(plVar15[1] + 0x30))(plVar15);
          }
          FUN_18000d1d0(param_1,plVar12);
          uVar26 = 0x11f;
          goto LAB_1802e08a9;
        }
        plVar15 = (longlong *)FUN_18037f0c0();
        if (plVar15 == (longlong *)0x0) {
          uVar26 = 0x11f;
          goto LAB_1802e0893;
        }
        plVar12 = (longlong *)
                  FUN_1803609a0(FUN_1802e1470,DAT_1803c5680,DAT_1803c5818,DAT_1803c53d8,0,0,0,
                                DAT_1803ca450,0,0,0);
        uVar26 = 0x11f;
        *(undefined4 *)(plVar6 + 5) = 0x11f;
        local_res18 = plVar12;
        plVar16 = (longlong *)FUN_18037ccb0(param_1,plVar15,&local_res18,DAT_1803c5690);
        if ((-1 < (int)*plVar15) && (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
          (**(code **)(plVar15[1] + 0x30))(plVar15);
        }
        if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
          (**(code **)(plVar12[1] + 0x30))(plVar12);
        }
        if (plVar16 == (longlong *)0x0) goto LAB_1802e0893;
        if ((-1 < (int)*plVar16) && (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
          (**(code **)(plVar16[1] + 0x30))(plVar16);
        }
        uVar18 = DAT_1803c54e8;
        plVar15 = (longlong *)_PyDict_NewPresized(2);
        PyDict_SetItem(plVar15,uVar18,plVar25);
        PyDict_SetItem(plVar15,DAT_1803c5508,plVar14);
        FUN_180001d80(param_1);
        goto LAB_1802e0bd3;
      }
      plVar15 = (longlong *)(*pcVar23)(plVar16);
      if (plVar15 == (longlong *)0x0) goto LAB_1802e065c;
      if (((local_b0 != (longlong *)0x0) && (-1 < (int)*local_b0)) &&
         (*local_b0 = *local_b0 + -1, *local_b0 == 0)) {
        (**(code **)(local_b0[1] + 0x30))(local_b0);
      }
      local_b0 = plVar15;
      if (*(code **)(plVar15[1] + 0xd8) == (code *)0x0) {
        iVar9 = PySequence_Check(plVar15);
        if (iVar9 == 0) {
          PyErr_Format(*(undefined8 *)PyExc_TypeError_exref,"cannot unpack non-iterable %s object",
                       *(undefined8 *)(plVar15[1] + 0x18));
        }
        else {
          plVar19 = (longlong *)FUN_180001220(PySeqIter_Type_exref);
          plVar19[2] = 0;
          iVar9 = (int)*plVar15 + 1;
          if (iVar9 != 0) {
            *(int *)plVar15 = iVar9;
          }
          plVar19[3] = (longlong)plVar15;
          plVar15 = plVar19 + -2;
          lVar17 = _PyThreadState_GetCurrent();
          lVar17 = *(longlong *)(*(longlong *)(lVar17 + 0x10) + 0xd0);
          plVar12 = *(longlong **)(lVar17 + 8);
          *plVar12 = (longlong)plVar15;
          plVar19[-1] = plVar19[-1] & 3;
          plVar19[-1] = plVar19[-1] | (ulonglong)plVar12;
          *plVar15 = lVar17;
          *(longlong **)(lVar17 + 8) = plVar15;
          if (plVar19 != (longlong *)0x0) goto LAB_1802df71b;
        }
LAB_1802e0626:
        uVar26 = 0x10e;
        plVar12 = *(longlong **)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
        goto LAB_1802e0203;
      }
      plVar19 = (longlong *)(**(code **)(plVar15[1] + 0xd8))();
      if (plVar19 == (longlong *)0x0) goto LAB_1802e0626;
      lVar17 = *(longlong *)(plVar19[1] + 0xe0);
      if ((lVar17 == 0) || (lVar17 == DAT_1803ca310)) {
        PyErr_Format(*(undefined8 *)PyExc_TypeError_exref,
                     "iter() returned non-iterator of type \'%s\'",
                     *(undefined8 *)(plVar19[1] + 0x18));
        if ((-1 < (int)*plVar19) && (*plVar19 = *plVar19 + -1, *plVar19 == 0)) {
          (**(code **)(plVar19[1] + 0x30))(plVar19);
        }
        goto LAB_1802e0626;
      }
LAB_1802df71b:
      plVar15 = (longlong *)(**(code **)(plVar19[1] + 0xe0))(plVar19);
      if (plVar15 == (longlong *)0x0) {
        plVar12 = *(longlong **)(param_1 + 0x60);
        if ((plVar12 == (longlong *)0x0) || (lVar17 = plVar12[1], lVar17 == 0)) {
LAB_1802e0570:
          plVar15 = (longlong *)
                    PyUnicode_FromFormat("not enough values to unpack (expected %d, got %d)",2,0);
          plVar12 = (longlong *)FUN_180376770(param_1,*(undefined8 *)PyExc_ValueError_exref,plVar15)
          ;
          if ((-1 < (int)*plVar15) && (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
            (**(code **)(plVar15[1] + 0x30))(plVar15);
          }
        }
        else {
          if ((*(uint *)(*(longlong *)(lVar17 + 8) + 0xa8) & 0x40000000) != 0) {
            lVar17 = *(longlong *)(lVar17 + 8);
          }
          plVar15 = plVar12;
          if ((lVar17 == *(longlong *)PyExc_StopIteration_exref) ||
             (((*(int *)(*(longlong *)(lVar17 + 8) + 0xa8) < 0 &&
               ((*(uint *)(lVar17 + 0xa8) & 0x40000000) != 0)) &&
              (cVar8 = FUN_180382640(), cVar8 == '\x01')))) {
            *(undefined8 *)(param_1 + 0x60) = 0;
            if ((-1 < (int)*plVar15) && (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
              (**(code **)(plVar15[1] + 0x30))(plVar15);
            }
            goto LAB_1802e0570;
          }
          *(undefined8 *)(param_1 + 0x60) = 0;
        }
        plVar15 = (longlong *)0x0;
LAB_1802e05bb:
        plVar22 = (longlong *)0x0;
LAB_1802e05bd:
        uVar26 = 0x10e;
        if ((-1 < (int)*plVar19) && (*plVar19 = *plVar19 + -1, *plVar19 == 0)) {
          (**(code **)(plVar19[1] + 0x30))(plVar19);
        }
        if (((plVar15 != (longlong *)0x0) && (-1 < (int)*plVar15)) &&
           (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
          (**(code **)(plVar15[1] + 0x30))(plVar15);
        }
        if (plVar22 != (longlong *)0x0) {
LAB_1802e02de:
          if ((-1 < (int)*plVar22) && (*plVar22 = *plVar22 + -1, *plVar22 == 0)) {
            (**(code **)(plVar22[1] + 0x30))(plVar22);
          }
        }
        goto LAB_1802e0203;
      }
      plVar20 = (longlong *)(**(code **)(plVar19[1] + 0xe0))(plVar19);
      if (plVar20 == (longlong *)0x0) {
        plVar12 = *(longlong **)(param_1 + 0x60);
        if ((plVar12 == (longlong *)0x0) || (lVar17 = plVar12[1], lVar17 == 0)) {
LAB_1802e048c:
          plVar25 = (longlong *)
                    PyUnicode_FromFormat("not enough values to unpack (expected %d, got %d)",2,1);
          plVar12 = (longlong *)FUN_180376770(param_1,*(undefined8 *)PyExc_ValueError_exref,plVar25)
          ;
          if ((-1 < (int)*plVar25) && (*plVar25 = *plVar25 + -1, *plVar25 == 0)) {
            (**(code **)(plVar25[1] + 0x30))(plVar25);
          }
        }
        else {
          if ((*(uint *)(*(longlong *)(lVar17 + 8) + 0xa8) & 0x40000000) != 0) {
            lVar17 = *(longlong *)(lVar17 + 8);
          }
          plVar25 = plVar12;
          if ((lVar17 == *(longlong *)PyExc_StopIteration_exref) ||
             (((*(int *)(*(longlong *)(lVar17 + 8) + 0xa8) < 0 &&
               ((*(uint *)(lVar17 + 0xa8) & 0x40000000) != 0)) &&
              (cVar8 = FUN_180382640(), cVar8 == '\x01')))) {
            *(undefined8 *)(param_1 + 0x60) = 0;
            if ((-1 < (int)*plVar25) && (*plVar25 = *plVar25 + -1, *plVar25 == 0)) {
              (**(code **)(plVar25[1] + 0x30))(plVar25);
            }
            goto LAB_1802e048c;
          }
          *(undefined8 *)(param_1 + 0x60) = 0;
        }
        goto LAB_1802e05bb;
      }
      plVar12 = (longlong *)(**(code **)(plVar19[1] + 0xe0))(plVar19);
      plVar22 = plVar20;
      if (plVar12 != (longlong *)0x0) {
        if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
          (**(code **)(plVar12[1] + 0x30))(plVar12);
        }
        plVar25 = (longlong *)PyUnicode_FromFormat("too many values to unpack (expected %d)",2);
        plVar12 = (longlong *)FUN_180376770(param_1,*(undefined8 *)PyExc_ValueError_exref,plVar25);
        if ((-1 < (int)*plVar25) && (*plVar25 = *plVar25 + -1, *plVar25 == 0)) {
          (**(code **)(plVar25[1] + 0x30))(plVar25);
        }
        goto LAB_1802e05bd;
      }
      plVar12 = *(longlong **)(param_1 + 0x60);
      if ((plVar12 != (longlong *)0x0) && (lVar17 = plVar12[1], lVar17 != 0)) {
        if ((*(uint *)(*(longlong *)(lVar17 + 8) + 0xa8) & 0x40000000) != 0) {
          lVar17 = *(longlong *)(lVar17 + 8);
        }
        if (lVar17 == *(longlong *)PyExc_StopIteration_exref) {
          *(undefined8 *)(param_1 + 0x60) = 0;
LAB_1802df7eb:
          if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
            (**(code **)(plVar12[1] + 0x30))(plVar12);
          }
          goto LAB_1802df803;
        }
        if ((*(int *)(*(longlong *)(lVar17 + 8) + 0xa8) < 0) &&
           ((*(uint *)(lVar17 + 0xa8) & 0x40000000) != 0)) {
          cVar8 = FUN_180382640();
          *(undefined8 *)(param_1 + 0x60) = 0;
          if (cVar8 == '\x01') goto LAB_1802df7eb;
        }
        else {
          *(undefined8 *)(param_1 + 0x60) = 0;
        }
        goto LAB_1802e05bd;
      }
LAB_1802df803:
      if ((-1 < (int)*plVar19) && (*plVar19 = *plVar19 + -1, *plVar19 == 0)) {
        (**(code **)(plVar19[1] + 0x30))(plVar19);
      }
      iVar9 = (int)*plVar15 + 1;
      if (iVar9 != 0) {
        *(int *)plVar15 = iVar9;
      }
      if (((local_d0 != (longlong *)0x0) && (-1 < (int)*local_d0)) &&
         (*local_d0 = *local_d0 + -1, *local_d0 == 0)) {
        (**(code **)(local_d0[1] + 0x30))();
      }
      if ((-1 < (int)*plVar15) && (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
        (**(code **)(plVar15[1] + 0x30))(plVar15);
      }
      iVar9 = (int)*plVar20 + 1;
      if (iVar9 != 0) {
        *(int *)plVar20 = iVar9;
      }
      if (((local_b8 != (longlong *)0x0) && (-1 < (int)*local_b8)) &&
         (*local_b8 = *local_b8 + -1, *local_b8 == 0)) {
        (**(code **)(local_b8[1] + 0x30))();
      }
      if ((-1 < (int)*plVar20) && (*plVar20 = *plVar20 + -1, *plVar20 == 0)) {
        (**(code **)(plVar20[1] + 0x30))(plVar20);
      }
      plVar12 = (longlong *)FUN_18037f0c0();
      local_d0 = plVar15;
      local_b8 = plVar20;
      if (plVar12 == (longlong *)0x0) {
LAB_1802e037b:
        uVar26 = 0x10f;
        plVar12 = *(longlong **)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
        goto LAB_1802e0203;
      }
      *(undefined4 *)(plVar6 + 5) = 0x10f;
      pcVar23 = (code *)FUN_180376770(param_1,plVar12,plVar15);
      if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
        (**(code **)(plVar12[1] + 0x30))(plVar12);
      }
      if (pcVar23 == (code *)0x0) goto LAB_1802e037b;
      if (((local_c8 != (code *)0x0) && (-1 < *(int *)local_c8)) &&
         (*(longlong *)local_c8 = *(longlong *)local_c8 + -1, *(longlong *)local_c8 == 0)) {
        (**(code **)(*(longlong *)(local_c8 + 8) + 0x30))(local_c8);
      }
      local_c8 = pcVar23;
    } while (pcVar23 == _Py_NoneStruct_exref);
    if (plVar14 == (longlong *)0x0) {
      plVar15 = (longlong *)
                PyUnicode_FromFormat
                          ("cannot access local variable \'%U\' where it is not associated with a value"
                           ,DAT_1803c57b0);
      plVar12 = (longlong *)
                FUN_180376770(param_1,*(undefined8 *)PyExc_UnboundLocalError_exref,plVar15);
      if ((-1 < (int)*plVar15) && (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
        (**(code **)(plVar15[1] + 0x30))(plVar15);
      }
      FUN_18000d1d0(param_1,plVar12);
      uVar26 = 0x112;
      goto LAB_1802e0203;
    }
    lVar17 = FUN_180373860(plVar14);
    if (((lVar17 < 0) && (*(longlong *)(param_1 + 0x60) != 0)) ||
       (plVar15 = (longlong *)PyLong_FromSsize_t(lVar17), plVar15 == (longlong *)0x0)) {
LAB_1802e0303:
      plVar12 = *(longlong **)(param_1 + 0x60);
      uVar26 = 0x112;
      *(undefined8 *)(param_1 + 0x60) = 0;
      goto LAB_1802e0203;
    }
    uVar26 = plVar15[2];
    local_res20[0] = 1;
    if (uVar26 >> 3 < 2) {
      plVar12 = (longlong *)FUN_1803800f0((1 - ((uint)uVar26 & 3)) * (int)plVar15[3] + -1);
    }
    else if ((uVar26 & 2) == 0) {
      plVar12 = (longlong *)FUN_180380640(plVar15 + 3,uVar26 >> 3,local_res20,1);
    }
    else {
      plVar12 = (longlong *)FUN_180380380();
      plVar12[2] = (ulonglong)(2 - (*(uint *)(plVar12 + 2) & 3)) | plVar12[2] & 0xfffffffffffffff8U;
    }
    if ((-1 < (int)*plVar15) && (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
      (**(code **)(plVar15[1] + 0x30))(plVar15);
    }
    plVar15 = (longlong *)FUN_180375000(param_1,plVar12,DAT_1803ab260,DAT_1803ab260);
    if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
      (**(code **)(plVar12[1] + 0x30))(plVar12);
    }
    if (plVar15 == (longlong *)0x0) goto LAB_1802e0303;
    plVar19 = (longlong *)FUN_180001930();
    if ((-1 < (int)*plVar15) && (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
      (**(code **)(plVar15[1] + 0x30))(plVar15);
    }
    if (plVar19 == (longlong *)0x0) goto LAB_1802e0303;
    plVar12 = (longlong *)(**(code **)(plVar19[1] + 0xe0))(plVar19);
    plVar15 = (longlong *)0x0;
    while (plVar22 = plVar12, plVar22 != (longlong *)0x0) {
      if (((plVar15 != (longlong *)0x0) && (-1 < (int)*plVar15)) &&
         (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
        (**(code **)(plVar15[1] + 0x30))();
      }
      iVar9 = (int)*plVar22 + 1;
      if (iVar9 != 0) {
        *(int *)plVar22 = iVar9;
      }
      if (((local_e0 != (longlong *)0x0) && (-1 < (int)*local_e0)) &&
         (*local_e0 = *local_e0 + -1, *local_e0 == 0)) {
        (**(code **)(local_e0[1] + 0x30))();
      }
      plVar15 = (longlong *)FUN_180024e50(param_1,plVar14,plVar22);
      local_e0 = plVar22;
      if (plVar15 == (longlong *)0x0) {
LAB_1802e01a8:
        plVar12 = *(longlong **)(param_1 + 0x60);
        uVar10 = local_res18._0_4_;
        goto LAB_1802e01b6;
      }
      plVar12 = (longlong *)FUN_18037f0c0();
      if ((-1 < (int)*plVar15) && (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
        (**(code **)(plVar15[1] + 0x30))(plVar15);
      }
      if (plVar12 == (longlong *)0x0) goto LAB_1802e01a8;
      *(undefined4 *)(plVar6 + 5) = 0x113;
      pcVar21 = (code *)FUN_1803770d0(param_1,plVar12,DAT_1803c57b8);
      if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
        (**(code **)(plVar12[1] + 0x30))(plVar12);
      }
      if (pcVar21 == (code *)0x0) goto LAB_1802e01a8;
      if (pcVar21 != _Py_TrueStruct_exref) {
        if ((pcVar21 != _Py_FalseStruct_exref) && (pcVar21 != _Py_NoneStruct_exref)) {
          lVar17 = *(longlong *)(pcVar21 + 8);
          if ((*(longlong *)(lVar17 + 0x60) == 0) ||
             (pcVar24 = *(code **)(*(longlong *)(lVar17 + 0x60) + 0x48), pcVar24 == (code *)0x0)) {
            if (((*(longlong **)(lVar17 + 0x70) == (longlong *)0x0) ||
                (pcVar24 = (code *)**(longlong **)(lVar17 + 0x70), pcVar24 == (code *)0x0)) &&
               ((*(longlong **)(lVar17 + 0x68) == (longlong *)0x0 ||
                (pcVar24 = (code *)**(longlong **)(lVar17 + 0x68), pcVar24 == (code *)0x0))))
            goto LAB_1802dfc22;
            lVar17 = (*pcVar24)(pcVar21);
          }
          else {
            iVar9 = (*pcVar24)(pcVar21);
            lVar17 = (longlong)iVar9;
          }
          if (0 < lVar17) goto LAB_1802dfc22;
          if (lVar17 != 0) {
            plVar12 = *(longlong **)(param_1 + 0x60);
            *(undefined8 *)(param_1 + 0x60) = 0;
            if ((-1 < *(int *)pcVar21) &&
               (*(longlong *)pcVar21 = *(longlong *)pcVar21 + -1, *(longlong *)pcVar21 == 0)) {
              (**(code **)(*(longlong *)(pcVar21 + 8) + 0x30))(pcVar21);
            }
            goto LAB_1802e01c1;
          }
        }
        if ((-1 < *(int *)pcVar21) &&
           (*(longlong *)pcVar21 = *(longlong *)pcVar21 + -1, *(longlong *)pcVar21 == 0)) {
          (**(code **)(*(longlong *)(pcVar21 + 8) + 0x30))(pcVar21);
        }
        pcVar21 = DAT_1803c5550;
        if (*(int *)DAT_1803c5550 + 1 != 0) {
          *(int *)DAT_1803c5550 = *(int *)DAT_1803c5550 + 1;
        }
      }
LAB_1802dfc22:
      lVar17 = FUN_1803749a0(pcVar21);
      if ((-1 < *(int *)pcVar21) &&
         (*(longlong *)pcVar21 = *(longlong *)pcVar21 + -1, *(longlong *)pcVar21 == 0)) {
        (**(code **)(*(longlong *)(pcVar21 + 8) + 0x30))(pcVar21);
      }
      if (lVar17 == 0) goto LAB_1802e01a8;
      plVar15 = (longlong *)plVar11[2];
      plVar11[2] = lVar17;
      if (((plVar15 != (longlong *)0x0) && (-1 < (int)*plVar15)) &&
         (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
        (**(code **)(plVar15[1] + 0x30))();
      }
      plVar15 = (longlong *)FUN_180001930();
      if (plVar15 == (longlong *)0x0) {
        plVar12 = *(longlong **)(param_1 + 0x60);
        uVar10 = 0x114;
        goto LAB_1802e01b6;
      }
      local_50 = FUN_18036d750(plVar15);
      iVar9 = (int)*plVar11 + 1;
      if (iVar9 != 0) {
        *(int *)plVar11 = iVar9;
      }
      local_48 = plVar11;
      plVar12 = (longlong *)
                FUN_180367890(FUN_1802e0d50,DAT_1803ca450,DAT_1803c55b8,DAT_1803c5820,DAT_1803c53f0,
                              &local_50,2,0x448);
      if ((-1 < (int)*plVar15) && (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
        (**(code **)(plVar15[1] + 0x30))(plVar15);
      }
      plVar15 = (longlong *)FUN_1803753b0(param_1,plVar12);
      if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
        (**(code **)(plVar12[1] + 0x30))(plVar12);
      }
      if (plVar15 == (longlong *)0x0) {
        plVar12 = *(longlong **)(param_1 + 0x60);
        local_res18 = (longlong *)CONCAT44(local_res18._4_4_,0x114);
        goto LAB_1802e01b9;
      }
      iVar9 = FUN_180001510(plVar15);
      if ((-1 < (int)*plVar15) && (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
        (**(code **)(plVar15[1] + 0x30))(plVar15);
      }
      if (iVar9 == 1) {
        plVar15 = (longlong *)FUN_18037f0c0();
        if (plVar15 != (longlong *)0x0) {
          *(undefined4 *)(plVar6 + 5) = 0x115;
          plVar12 = (longlong *)FUN_180376770(param_1,plVar15,plVar22);
          if ((-1 < (int)*plVar15) && (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
            (**(code **)(plVar15[1] + 0x30))(plVar15);
          }
          if (plVar12 != (longlong *)0x0) {
            if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
              (**(code **)(plVar12[1] + 0x30))(plVar12);
            }
            goto LAB_1802dfdd1;
          }
        }
        plVar12 = *(longlong **)(param_1 + 0x60);
        local_res18 = (longlong *)CONCAT44(local_res18._4_4_,0x115);
        goto LAB_1802e01b9;
      }
LAB_1802dfdd1:
      lVar17 = *(longlong *)(param_1 + 0x10);
      if ((((*(int *)(*(longlong *)(lVar17 + 0x60) + 0x4d4) != 0) || (*(int *)(lVar17 + 0x1a0) != 0)
           ) && (iVar9 = Py_MakePendingCalls(), iVar9 < 0)) &&
         (plVar12 = *(longlong **)(param_1 + 0x60), plVar12 != (longlong *)0x0)) {
LAB_1802e0187:
        local_res18 = (longlong *)CONCAT44(local_res18._4_4_,0x112);
        goto LAB_1802e01b9;
      }
      if (*(int *)(lVar17 + 0x174) != 0) {
        PyEval_SaveThread();
        PyEval_AcquireThread(param_1);
      }
      if (*(longlong *)(param_1 + 0x80) != 0) {
        plVar15 = *(longlong **)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x80) = 0;
        plVar12 = (longlong *)FUN_180373490(param_1);
        *(longlong **)(param_1 + 0x60) = plVar12;
        if ((plVar15 != (longlong *)0x0) && (-1 < (int)*plVar15)) {
          *plVar15 = *plVar15 + -1;
          if (*plVar15 == 0) {
            (**(code **)(plVar15[1] + 0x30))(plVar15);
          }
          plVar12 = *(longlong **)(param_1 + 0x60);
        }
        goto LAB_1802e0187;
      }
      plVar12 = (longlong *)(**(code **)(plVar19[1] + 0xe0))(plVar19);
      plVar15 = plVar22;
    }
    plVar12 = *(longlong **)(param_1 + 0x60);
    if (plVar12 != (longlong *)0x0) {
      plVar22 = plVar12;
      if ((*(uint *)((longlong *)plVar12[1] + 0x15) & 0x40000000) != 0) {
        plVar22 = (longlong *)plVar12[1];
      }
      if ((plVar22 != *(longlong **)PyExc_StopIteration_exref) &&
         (((-1 < *(int *)(plVar22[1] + 0xa8) || ((*(uint *)(plVar22 + 0x15) & 0x40000000) == 0)) ||
          (cVar8 = FUN_180382640(), cVar8 != '\x01')))) {
        plVar22 = plVar15;
        uVar10 = 0x112;
LAB_1802e01b6:
        local_res18 = (longlong *)CONCAT44(local_res18._4_4_,uVar10);
LAB_1802e01b9:
        *(undefined8 *)(param_1 + 0x60) = 0;
LAB_1802e01c1:
        if (((plVar22 != (longlong *)0x0) && (-1 < (int)*plVar22)) &&
           (*plVar22 = *plVar22 + -1, *plVar22 == 0)) {
          (**(code **)(plVar22[1] + 0x30))(plVar22);
        }
        if ((-1 < (int)*plVar19) && (*plVar19 = *plVar19 + -1, *plVar19 == 0)) {
          (**(code **)(plVar19[1] + 0x30))(plVar19);
        }
        uVar26 = (ulonglong)local_res18 & 0xffffffff;
        goto LAB_1802e0203;
      }
      *(undefined8 *)(param_1 + 0x60) = 0;
      if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
        (**(code **)(plVar12[1] + 0x30))(plVar12);
      }
    }
    if (((plVar15 != (longlong *)0x0) && (-1 < (int)*plVar15)) &&
       (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
      (**(code **)(plVar15[1] + 0x30))(plVar15);
    }
    if ((-1 < (int)*plVar19) && (*plVar19 = *plVar19 + -1, *plVar19 == 0)) {
      (**(code **)(plVar19[1] + 0x30))(plVar19);
    }
    plVar15 = (longlong *)FUN_18037f0c0();
    uVar7 = DAT_1803c57e0;
    uVar18 = DAT_1803c57d8;
    if (plVar15 == (longlong *)0x0) {
LAB_1802e0297:
      plVar12 = *(longlong **)(param_1 + 0x60);
      uVar26 = 0x116;
      *(undefined8 *)(param_1 + 0x60) = 0;
      goto LAB_1802e0203;
    }
    plVar22 = (longlong *)_PyDict_NewPresized(6);
    PyDict_SetItem(plVar22,uVar18,uVar7);
    uVar18 = DAT_1803c57e8;
    plVar12 = (longlong *)FUN_180058050(param_1,plVar20,DAT_1803ab250);
    if (plVar12 == (longlong *)0x0) {
      uVar26 = 0x118;
LAB_1802e02bc:
      plVar12 = *(longlong **)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      if ((-1 < (int)*plVar15) && (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
        (**(code **)(plVar15[1] + 0x30))(plVar15);
      }
      goto LAB_1802e02de;
    }
    PyDict_SetItem(plVar22,uVar18,plVar12);
    if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
      (**(code **)(plVar12[1] + 0x30))(plVar12);
    }
    uVar18 = DAT_1803c57f0;
    plVar12 = (longlong *)FUN_1803749a0(pcVar23);
    if (plVar12 == (longlong *)0x0) {
      uVar26 = 0x119;
      goto LAB_1802e02bc;
    }
    PyDict_SetItem(plVar22,uVar18,plVar12);
    if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
      (**(code **)(plVar12[1] + 0x30))(plVar12);
    }
    PyDict_SetItem(plVar22,DAT_1803c57f8,plVar25);
    PyDict_SetItem(plVar22,DAT_1803c5800,_Py_NoneStruct_exref);
    PyDict_SetItem(plVar22,DAT_1803c5808,plVar25);
    *(undefined4 *)(plVar6 + 5) = 0x116;
    plVar12 = (longlong *)FUN_180376770(param_1,plVar15,plVar22);
    if ((-1 < (int)*plVar15) && (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
      (**(code **)(plVar15[1] + 0x30))(plVar15);
    }
    if ((-1 < (int)*plVar22) && (*plVar22 = *plVar22 + -1, *plVar22 == 0)) {
      (**(code **)(plVar22[1] + 0x30))(plVar22);
    }
    if (plVar12 == (longlong *)0x0) goto LAB_1802e0297;
    if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
      (**(code **)(plVar12[1] + 0x30))(plVar12);
    }
    lVar17 = *(longlong *)(param_1 + 0x10);
    if ((((*(int *)(*(longlong *)(lVar17 + 0x60) + 0x4d4) != 0) || (*(int *)(lVar17 + 0x1a0) != 0))
        && (iVar9 = Py_MakePendingCalls(), iVar9 < 0)) &&
       (plVar12 = *(longlong **)(param_1 + 0x60), plVar12 != (longlong *)0x0)) goto LAB_1802e0284;
    if (*(int *)(lVar17 + 0x174) != 0) {
      PyEval_SaveThread();
      PyEval_AcquireThread(param_1);
    }
  } while (*(longlong *)(param_1 + 0x80) == 0);
  plVar15 = *(longlong **)(param_1 + 0x60);
  *(undefined8 *)(param_1 + 0x80) = 0;
  plVar12 = (longlong *)FUN_180373490(param_1);
  *(longlong **)(param_1 + 0x60) = plVar12;
  if ((plVar15 != (longlong *)0x0) && (-1 < (int)*plVar15)) {
    *plVar15 = *plVar15 + -1;
    if (*plVar15 == 0) {
      (**(code **)(plVar15[1] + 0x30))(plVar15);
    }
    plVar12 = *(longlong **)(param_1 + 0x60);
  }
LAB_1802e0284:
  *(undefined8 *)(param_1 + 0x60) = 0;
  uVar26 = 0x10e;
LAB_1802e0203:
  if ((-1 < (int)*local_b0) && (*local_b0 = *local_b0 + -1, *local_b0 == 0)) {
    (**(code **)(local_b0[1] + 0x30))(local_b0);
  }
LAB_1802e021f:
  if ((-1 < (int)*plVar16) && (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
    (**(code **)(plVar16[1] + 0x30))(plVar16);
  }
LAB_1802e08a9:
  piVar5 = (int *)plVar12[5];
  if (piVar5 == (int *)0x0) {
    pcVar23 = (code *)FUN_18038e2a0(plVar6,uVar26);
    plVar15 = (longlong *)plVar12[5];
    pcVar21 = pcVar23;
    if (pcVar23 == _Py_NoneStruct_exref) {
      pcVar21 = (code *)0x0;
    }
    else {
LAB_1802e0907:
      pcVar23 = pcVar21;
      if ((pcVar21 != (code *)0x0) && (*(int *)pcVar21 + 1 != 0)) {
        *(int *)pcVar21 = *(int *)pcVar21 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar5 + 6) == plVar6) goto LAB_1802e0950;
    pcVar23 = (code *)FUN_18038e2a0(plVar6,uVar26);
    *(int **)(pcVar23 + 0x10) = piVar5;
    if (*piVar5 + 1 != 0) {
      *piVar5 = *piVar5 + 1;
    }
    plVar15 = (longlong *)plVar12[5];
    pcVar21 = pcVar23;
    if (pcVar23 != _Py_NoneStruct_exref) goto LAB_1802e0907;
    pcVar21 = (code *)0x0;
  }
  plVar12[5] = (longlong)pcVar21;
  if (((plVar15 != (longlong *)0x0) && (-1 < (int)*plVar15)) &&
     (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
    (**(code **)(plVar15[1] + 0x30))();
  }
  if (((pcVar23 != (code *)0x0) && (-1 < *(int *)pcVar23)) &&
     (*(longlong *)pcVar23 = *(longlong *)pcVar23 + -1, *(longlong *)pcVar23 == 0)) {
    (**(code **)(*(longlong *)(pcVar23 + 8) + 0x30))(pcVar23);
  }
LAB_1802e0950:
  FUN_18036d0a0(plVar6,"oooooooooooc",plVar1,plVar3,plVar2,local_98,local_a8,local_88,local_80,
                local_d0,local_b8,local_c8,local_e0,plVar11);
  plVar15 = DAT_1803c53c0;
  if (plVar6 == DAT_1803c53c0) {
    if ((-1 < (int)*DAT_1803c53c0) && (*DAT_1803c53c0 = *DAT_1803c53c0 + -1, *plVar15 == 0)) {
      (**(code **)(plVar15[1] + 0x30))();
    }
    DAT_1803c53c0 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if (((local_98 != (longlong *)0x0) && (-1 < (int)*local_98)) &&
     (*local_98 = *local_98 + -1, *local_98 == 0)) {
    (**(code **)(local_98[1] + 0x30))();
  }
  if (((local_a8 != (longlong *)0x0) && (-1 < (int)*local_a8)) &&
     (*local_a8 = *local_a8 + -1, *local_a8 == 0)) {
    (**(code **)(local_a8[1] + 0x30))();
  }
  if (((local_88 != (longlong *)0x0) && (-1 < (int)*local_88)) &&
     (*local_88 = *local_88 + -1, *local_88 == 0)) {
    (**(code **)(local_88[1] + 0x30))();
  }
  if (((local_80 != (longlong *)0x0) && (-1 < (int)*local_80)) &&
     (*local_80 = *local_80 + -1, *local_80 == 0)) {
    (**(code **)(local_80[1] + 0x30))();
  }
  if (((local_d0 != (longlong *)0x0) && (-1 < (int)*local_d0)) &&
     (*local_d0 = *local_d0 + -1, *local_d0 == 0)) {
    (**(code **)(local_d0[1] + 0x30))();
  }
  if (((local_b8 != (longlong *)0x0) && (-1 < (int)*local_b8)) &&
     (*local_b8 = *local_b8 + -1, *local_b8 == 0)) {
    (**(code **)(local_b8[1] + 0x30))(local_b8);
  }
  if (((local_c8 != (code *)0x0) && (-1 < *(int *)local_c8)) &&
     (*(longlong *)local_c8 = *(longlong *)local_c8 + -1, *(longlong *)local_c8 == 0)) {
    (**(code **)(*(longlong *)(local_c8 + 8) + 0x30))(local_c8);
  }
  if (((local_e0 != (longlong *)0x0) && (-1 < (int)*local_e0)) &&
     (*local_e0 = *local_e0 + -1, *local_e0 == 0)) {
    (**(code **)(local_e0[1] + 0x30))(local_e0);
  }
  if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
    (**(code **)(plVar11[1] + 0x30))(plVar11);
  }
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  if ((-1 < (int)*plVar3) && (*plVar3 = *plVar3 + -1, *plVar3 == 0)) {
    (**(code **)(plVar3[1] + 0x30))();
  }
  if ((-1 < (int)*plVar2) && (*plVar2 = *plVar2 + -1, *plVar2 == 0)) {
    (**(code **)(plVar2[1] + 0x30))();
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong **)(param_1 + 0x60) = plVar12;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return (longlong *)0x0;
}

