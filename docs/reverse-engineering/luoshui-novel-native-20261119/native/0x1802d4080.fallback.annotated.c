/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x1802d4080 */

/* WARNING: Function: __security_check_cookie replaced with injection: security_check_cookie */

undefined8 FUN_1802d4080(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong lVar1;
  int *piVar2;
  char cVar3;
  int iVar4;
  longlong *plVar5;
  longlong *plVar6;
  longlong *plVar7;
  longlong *plVar8;
  longlong *plVar9;
  longlong lVar10;
  undefined8 uVar11;
  longlong *plVar12;
  undefined8 uVar13;
  code *pcVar14;
  code *pcVar15;
  undefined4 uVar16;
  bool bVar17;
  undefined1 auStack_1a8 [32];
  longlong *local_188;
  longlong **local_180;
  longlong *local_178;
  longlong *local_170;
  longlong *local_168;
  longlong *local_160;
  longlong *local_158;
  longlong *local_150;
  longlong *local_148;
  longlong *local_138;
  longlong *local_130;
  longlong *local_128;
  longlong *local_120;
  longlong *local_118;
  longlong *local_110;
  longlong *local_108;
  longlong *local_100;
  longlong *local_f8;
  longlong *local_f0;
  longlong *local_e8;
  longlong *local_e0;
  longlong *local_d8;
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
  ulonglong local_40;
  
  local_40 = DAT_1803a6040 ^ (ulonglong)auStack_1a8;
  local_128 = (longlong *)param_3[1];
  plVar8 = (longlong *)*param_3;
  local_c8 = (longlong *)param_3[2];
  local_f8 = (longlong *)0x0;
  local_98 = (longlong *)0x0;
  local_b0 = (longlong *)0x0;
  local_138 = (longlong *)0x0;
  local_100 = (longlong *)0x0;
  local_90 = (longlong *)0x0;
  local_88 = (longlong *)0x0;
  local_a0 = plVar8;
  local_a8 = (longlong *)FUN_18036d6d0();
  plVar7 = DAT_1803c53f8;
  local_d0 = DAT_180391ea8;
  local_f0 = (longlong *)0x0;
  local_e8 = (longlong *)0x0;
  local_d8 = (longlong *)0x0;
  local_c0 = (longlong *)0x0;
  local_b8 = (longlong *)0x0;
  local_e0 = (longlong *)0x0;
  local_120 = DAT_1803c53f8;
  if (DAT_1803c53f8 == (longlong *)0x0) {
LAB_1802d414d:
    DAT_1803c53f8 = (longlong *)FUN_18036cd10(param_1,DAT_1803c5410,DAT_1803ca450,0x58);
  }
  else if ((1 < *DAT_1803c53f8) || (DAT_1803c53f8[2] != 0)) {
    if ((DAT_1803c53f8 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803c53f8 && (*DAT_1803c53f8 = *DAT_1803c53f8 + -1, *plVar7 == 0)))) {
      (**(code **)(plVar7[1] + 0x30))(plVar7);
    }
    goto LAB_1802d414d;
  }
  plVar7 = DAT_1803c53f8;
  local_120 = DAT_1803c53f8;
  FUN_180001d00(param_1,DAT_1803c53f8);
/* novels.services.state_reducer#0="manifest" */
  plVar5 = (longlong *)FUN_180024e50(param_1,plVar8,DAT_1803c54e0);
  if (plVar5 == (longlong *)0x0) {
LAB_1802d6052:
    uVar16 = 0x13;
  }
  else {
    plVar6 = (longlong *)FUN_180370dc0(param_1,plVar5);
    local_130 = plVar6;
    if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
      (**(code **)(plVar5[1] + 0x30))(plVar5);
    }
/* novels.services.state_reducer#1="chapter" */
    uVar11 = DAT_1803c54e8;
    if (plVar6 == (longlong *)0x0) goto LAB_1802d6052;
    local_f8 = plVar6;
/* novels.services.state_reducer#2="currentState" */
    plVar7 = (longlong *)FUN_180024e50(param_1,plVar8,DAT_1803c54f0);
    if (plVar7 == (longlong *)0x0) {
      uVar16 = 0x15;
      plVar7 = local_120;
    }
    else {
      plVar5 = (longlong *)FUN_18037f0c0();
      if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
        (**(code **)(plVar7[1] + 0x30))(plVar7);
      }
      uVar16 = 0x15;
      plVar7 = local_120;
      if (plVar5 != (longlong *)0x0) {
        *(undefined4 *)(local_120 + 5) = 0x15;
        pcVar14 = (code *)FUN_1803770d0(param_1,plVar5,DAT_1803c5500);
        if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
          (**(code **)(plVar5[1] + 0x30))(plVar5);
        }
        if (pcVar14 == (code *)0x0) {
LAB_1802d4289:
          plVar9 = *(longlong **)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          plVar7 = local_120;
          plVar5 = local_128;
          plVar8 = local_138;
        }
        else {
          iVar4 = FUN_180001510(pcVar14);
          if (iVar4 == -1) {
            plVar9 = *(longlong **)(param_1 + 0x60);
            *(undefined8 *)(param_1 + 0x60) = 0;
            plVar7 = local_120;
            plVar5 = local_128;
            plVar8 = local_138;
            if ((-1 < *(int *)pcVar14) &&
               (*(longlong *)pcVar14 = *(longlong *)pcVar14 + -1, *(longlong *)pcVar14 == 0)) {
              (**(code **)(*(longlong *)(pcVar14 + 8) + 0x30))(pcVar14);
              plVar7 = local_120;
              plVar5 = local_128;
              plVar8 = local_138;
            }
          }
          else {
            if (iVar4 != 1) {
              if ((-1 < *(int *)pcVar14) &&
                 (*(longlong *)pcVar14 = *(longlong *)pcVar14 + -1, *(longlong *)pcVar14 == 0)) {
                (**(code **)(*(longlong *)(pcVar14 + 8) + 0x30))(pcVar14);
              }
              pcVar14 = DAT_1803ab250;
              if (*(int *)DAT_1803ab250 + 1 != 0) {
                *(int *)DAT_1803ab250 = *(int *)DAT_1803ab250 + 1;
              }
            }
            plVar7 = (longlong *)PyNumber_Long(pcVar14);
            if ((-1 < *(int *)pcVar14) &&
               (*(longlong *)pcVar14 = *(longlong *)pcVar14 + -1, *(longlong *)pcVar14 == 0)) {
              (**(code **)(*(longlong *)(pcVar14 + 8) + 0x30))(pcVar14);
            }
            if (plVar7 == (longlong *)0x0) {
              plVar9 = *(longlong **)(param_1 + 0x60);
              *(undefined8 *)(param_1 + 0x60) = 0;
              plVar7 = local_120;
              plVar5 = local_128;
              plVar8 = local_138;
            }
            else {
              local_80 = (longlong *)_PyDict_NewPresized(2);
              PyDict_SetItem(local_80,uVar11,plVar7);
              if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                (**(code **)(plVar7[1] + 0x30))(plVar7);
              }
/* novels.services.state_reducer#5="facts" */
              local_118 = DAT_1803c5508;
/* novels.services.state_reducer#2="currentState" */
              plVar8 = (longlong *)FUN_180024e50(param_1,plVar8,DAT_1803c54f0);
              uVar16 = 0x16;
              if (plVar8 == (longlong *)0x0) {
                plVar9 = *(longlong **)(param_1 + 0x60);
                *(undefined8 *)(param_1 + 0x60) = 0;
                plVar7 = local_120;
              }
              else {
                plVar5 = (longlong *)FUN_18037f0c0();
                if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                  (**(code **)(plVar8[1] + 0x30))(plVar8);
                }
                plVar7 = local_120;
                if (plVar5 == (longlong *)0x0) {
                  plVar9 = *(longlong **)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                }
                else {
                  *(undefined4 *)(local_120 + 5) = 0x16;
                  plVar8 = (longlong *)FUN_1803770d0(param_1,plVar5,DAT_1803c5510);
                  if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                    (**(code **)(plVar5[1] + 0x30))(plVar5);
                  }
                  if (plVar8 == (longlong *)0x0) {
                    plVar9 = *(longlong **)(param_1 + 0x60);
                    *(undefined8 *)(param_1 + 0x60) = 0;
                  }
                  else {
                    iVar4 = FUN_180001510(plVar8);
                    if (iVar4 == -1) {
                      plVar9 = *(longlong **)(param_1 + 0x60);
                      *(undefined8 *)(param_1 + 0x60) = 0;
                      if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                        (**(code **)(plVar8[1] + 0x30))(plVar8);
                      }
                    }
                    else {
                      if (iVar4 != 1) {
                        if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                          (**(code **)(plVar8[1] + 0x30))(plVar8);
                        }
                        plVar8 = (longlong *)FUN_180372160(param_1,0);
                      }
                      plVar5 = (longlong *)FUN_180001930();
                      if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                        (**(code **)(plVar8[1] + 0x30))(plVar8);
                      }
                      if (plVar5 == (longlong *)0x0) {
                        plVar9 = *(longlong **)(param_1 + 0x60);
                        *(undefined8 *)(param_1 + 0x60) = 0;
                      }
                      else {
                        plVar8 = (longlong *)FUN_180372160(param_1,0);
                        do {
                          plVar6 = local_c0;
                          plVar7 = local_f0;
                          pcVar14 = *(code **)(plVar5[1] + 0xe0);
                          if (pcVar14 == (code *)0x0) {
                            PyErr_Format(*(undefined8 *)PyExc_TypeError_exref,
                                         "\'%s\' object is not an iterator",
                                         *(undefined8 *)(plVar5[1] + 0x18));
LAB_1802d4688:
                            plVar9 = *(longlong **)(param_1 + 0x60);
                            if (plVar9 != (longlong *)0x0) {
                              plVar12 = (longlong *)plVar9[1];
                              if ((*(uint *)((longlong *)plVar9[1] + 0x15) & 0x40000000) == 0) {
                                plVar12 = plVar9;
                              }
                              if ((plVar12 != *(longlong **)PyExc_StopIteration_exref) &&
                                 (((-1 < *(int *)(plVar12[1] + 0xa8) ||
                                   ((*(uint *)(plVar12 + 0x15) & 0x40000000) == 0)) ||
                                  (cVar3 = FUN_180382640(), cVar3 != '\x01')))) goto LAB_1802d46e0;
                              *(undefined8 *)(param_1 + 0x60) = 0;
                              if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                (**(code **)(plVar9[1] + 0x30))(plVar9);
                              }
                            }
                            iVar4 = (int)*plVar8 + 1;
                            if (iVar4 != 0) {
                              *(int *)plVar8 = iVar4;
                            }
                            if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                              (**(code **)(plVar5[1] + 0x30))(plVar5);
                            }
                            if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                              (**(code **)(plVar8[1] + 0x30))(plVar8);
                            }
                            if (((plVar6 != (longlong *)0x0) && (-1 < (int)*plVar6)) &&
                               (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                              (**(code **)(plVar6[1] + 0x30))(plVar6);
                            }
                            if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) &&
                               (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                              (**(code **)(plVar7[1] + 0x30))(plVar7);
                            }
                            plVar7 = local_80;
                            PyDict_SetItem(local_80,local_118,plVar8);
                            if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                              (**(code **)(plVar8[1] + 0x30))(plVar8);
                            }
                            local_98 = plVar7;
/* novels.services.state_reducer#7="hooks" */
                            local_f0 = DAT_1803c5518;
/* novels.services.state_reducer#7="hooks" */
                            plVar8 = (longlong *)FUN_180024e50(param_1,local_a0,DAT_1803c5518);
                            uVar16 = 0x19;
                            if (plVar8 == (longlong *)0x0) goto LAB_1802d4289;
                            plVar5 = (longlong *)FUN_18037f0c0();
                            if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                              (**(code **)(plVar8[1] + 0x30))(plVar8);
                            }
                            plVar7 = local_120;
                            if (plVar5 == (longlong *)0x0) {
                              plVar9 = *(longlong **)(param_1 + 0x60);
                              *(undefined8 *)(param_1 + 0x60) = 0;
                              plVar5 = local_128;
                              plVar8 = local_138;
                              goto LAB_1802d6063;
                            }
                            *(undefined4 *)(local_120 + 5) = 0x19;
                            plVar6 = (longlong *)FUN_1803770d0(param_1,plVar5,DAT_1803c5520);
                            if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                              (**(code **)(plVar5[1] + 0x30))(plVar5);
                            }
                            if (plVar6 == (longlong *)0x0) {
                              plVar9 = *(longlong **)(param_1 + 0x60);
                              *(undefined8 *)(param_1 + 0x60) = 0;
                              plVar5 = local_128;
                              plVar8 = local_138;
                              goto LAB_1802d6063;
                            }
                            iVar4 = FUN_180001510(plVar6);
                            if (iVar4 == -1) goto LAB_1802d4925;
                            if (iVar4 != 1) {
                              if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                (**(code **)(plVar6[1] + 0x30))(plVar6);
                              }
                              plVar6 = (longlong *)FUN_180372160(param_1,0);
                            }
                            plVar8 = (longlong *)FUN_180001930();
                            if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                              (**(code **)(plVar6[1] + 0x30))(plVar6);
                            }
                            if (plVar8 != (longlong *)0x0) {
                              plVar7 = (longlong *)FUN_180372160(param_1,0);
                              goto LAB_1802d49d0;
                            }
                            plVar9 = *(longlong **)(param_1 + 0x60);
                            *(undefined8 *)(param_1 + 0x60) = 0;
                            plVar5 = local_128;
                            plVar8 = local_138;
                            goto LAB_1802d6063;
                          }
                          plVar9 = (longlong *)(*pcVar14)(plVar5);
                          if (plVar9 == (longlong *)0x0) goto LAB_1802d4688;
                          local_c0 = plVar9;
                          if (((plVar6 != (longlong *)0x0) && (-1 < (int)*plVar6)) &&
                             (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                            (**(code **)(plVar6[1] + 0x30))(plVar6);
                          }
                          iVar4 = (int)*plVar9 + 1;
                          if (iVar4 != 0) {
                            *(int *)plVar9 = iVar4;
                          }
                          local_f0 = plVar9;
                          if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) &&
                             (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                            (**(code **)(plVar7[1] + 0x30))();
                          }
                          iVar4 = PyObject_IsInstance(plVar9,PyDict_Type_exref);
                          if (iVar4 == -1) goto LAB_1802d4662;
                          if (iVar4 != 0) {
                            lVar10 = FUN_180370dc0(param_1,plVar9);
                            if (lVar10 == 0) goto LAB_1802d4662;
                            lVar1 = plVar8[2];
                            cVar3 = FUN_1803722b0(plVar8,lVar1 + 1);
                            if (cVar3 == '\0') goto LAB_1802d4662;
                            *(longlong *)(plVar8[3] + lVar1 * 8) = lVar10;
                          }
                          lVar10 = *(longlong *)(param_1 + 0x10);
                          if (((*(int *)(*(longlong *)(lVar10 + 0x60) + 0x4d4) != 0) ||
                              (*(int *)(lVar10 + 0x1a0) != 0)) &&
                             ((iVar4 = Py_MakePendingCalls(), iVar4 < 0 &&
                              (*(longlong *)(param_1 + 0x60) != 0)))) goto LAB_1802d4662;
                          if (*(int *)(lVar10 + 0x174) != 0) {
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
LAB_1802d4662:
                        plVar9 = *(longlong **)(param_1 + 0x60);
LAB_1802d46e0:
                        *(undefined8 *)(param_1 + 0x60) = 0;
                        if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                          (**(code **)(plVar5[1] + 0x30))(plVar5);
                        }
                        if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                          (**(code **)(plVar8[1] + 0x30))(plVar8);
                        }
                        if (((local_c0 != (longlong *)0x0) && (-1 < (int)*local_c0)) &&
                           (*local_c0 = *local_c0 + -1, *local_c0 == 0)) {
                          (**(code **)(local_c0[1] + 0x30))();
                        }
                        plVar7 = local_120;
                        if (((local_f0 != (longlong *)0x0) && (-1 < (int)*local_f0)) &&
                           (*local_f0 = *local_f0 + -1, *local_f0 == 0)) {
                          (**(code **)(local_f0[1] + 0x30))();
                          plVar7 = local_120;
                        }
                      }
                    }
                  }
                }
              }
              plVar5 = local_128;
              plVar8 = local_138;
              if ((-1 < (int)*local_80) && (*local_80 = *local_80 + -1, *local_80 == 0)) {
                (**(code **)(local_80[1] + 0x30))();
                plVar5 = local_128;
                plVar8 = local_138;
              }
            }
          }
        }
        goto LAB_1802d6063;
      }
    }
  }
  plVar9 = *(longlong **)(param_1 + 0x60);
  *(undefined8 *)(param_1 + 0x60) = 0;
  plVar5 = local_128;
  plVar8 = (longlong *)0x0;
  goto LAB_1802d6063;
LAB_1802d4925:
  plVar9 = *(longlong **)(param_1 + 0x60);
  *(undefined8 *)(param_1 + 0x60) = 0;
  plVar5 = local_128;
  plVar8 = local_138;
  if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
    (**(code **)(plVar6[1] + 0x30))(plVar6);
    plVar5 = local_128;
    plVar8 = local_138;
  }
  goto LAB_1802d6063;
joined_r0x0001802d4e75:
  local_d8 = plVar5;
  if (plVar7 == (longlong *)0x0) goto LAB_1802d4f8d;
  plVar9 = plVar7;
  if (((local_e0 != (longlong *)0x0) && (-1 < (int)*local_e0)) &&
     (*local_e0 = *local_e0 + -1, *local_e0 == 0)) {
    plVar9 = local_e0 + 1;
    local_e0 = plVar7;
    (**(code **)(*plVar9 + 0x30))();
    plVar9 = local_e0;
  }
  local_e0 = plVar9;
  iVar4 = (int)*plVar7 + 1;
  if (iVar4 != 0) {
    *(int *)plVar7 = iVar4;
  }
  local_d8 = plVar7;
  if (((plVar5 != (longlong *)0x0) && (-1 < (int)*plVar5)) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)
     ) {
    (**(code **)(plVar5[1] + 0x30))();
  }
  iVar4 = PyObject_IsInstance(plVar7,PyDict_Type_exref);
  if ((iVar4 == -1) ||
     ((iVar4 != 0 &&
      ((lVar10 = FUN_180370dc0(param_1,plVar7), lVar10 == 0 ||
       (cVar3 = FUN_180372560(plVar6,lVar10), cVar3 == '\0')))))) {
LAB_1802d5100:
    plVar9 = *(longlong **)(param_1 + 0x60);
    goto LAB_1802d5104;
  }
  lVar10 = *(longlong *)(param_1 + 0x10);
  if (((*(int *)(*(longlong *)(lVar10 + 0x60) + 0x4d4) != 0) || (*(int *)(lVar10 + 0x1a0) != 0)) &&
     ((iVar4 = Py_MakePendingCalls(), iVar4 < 0 &&
      (plVar9 = *(longlong **)(param_1 + 0x60), plVar9 != (longlong *)0x0)))) goto LAB_1802d5104;
  if (*(int *)(lVar10 + 0x174) != 0) {
    PyEval_SaveThread();
    PyEval_AcquireThread(param_1);
  }
  if (*(longlong *)(param_1 + 0x80) != 0) {
    plVar7 = *(longlong **)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x80) = 0;
    plVar9 = (longlong *)FUN_180373490(param_1);
    *(longlong **)(param_1 + 0x60) = plVar9;
    if ((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) {
      *plVar7 = *plVar7 + -1;
      if (*plVar7 == 0) {
        (**(code **)(plVar7[1] + 0x30))(plVar7);
      }
      goto LAB_1802d5100;
    }
    goto LAB_1802d5104;
  }
  plVar7 = (longlong *)FUN_18000d430(plVar8);
  plVar5 = local_d8;
  goto joined_r0x0001802d4e75;
LAB_1802d4f8d:
  plVar9 = *(longlong **)(param_1 + 0x60);
  if (plVar9 != (longlong *)0x0) {
    cVar3 = FUN_1800014b0();
    if (cVar3 == '\0') {
LAB_1802d5104:
      *(undefined8 *)(param_1 + 0x60) = 0;
      if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
        (**(code **)(plVar8[1] + 0x30))(plVar8);
      }
      if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
        (**(code **)(plVar6[1] + 0x30))(plVar6);
      }
      plVar7 = local_120;
      plVar5 = local_128;
      plVar8 = local_138;
      if (local_d8 != (longlong *)0x0) {
        plVar6 = local_d8;
        if ((int)*local_d8 < 0) goto LAB_1802d4c25;
        *local_d8 = *local_d8 + -1;
        if (*local_d8 == 0) {
          (**(code **)(local_d8[1] + 0x30))();
          plVar6 = local_d8;
        }
        goto LAB_1802d4c2e;
      }
      goto LAB_1802d6063;
    }
    *(undefined8 *)(param_1 + 0x60) = 0;
    if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
      (**(code **)(plVar9[1] + 0x30))(plVar9);
    }
  }
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
  plVar8 = local_d8;
  if ((local_d8 != (longlong *)0x0) && (-1 < (int)*local_d8)) {
    *local_d8 = *local_d8 + -1;
    iVar4 = (int)*local_d8;
    if (*local_d8 == 0) {
      (**(code **)(local_d8[1] + 0x30))(local_d8);
      iVar4 = (int)*plVar8;
    }
    if ((-1 < iVar4) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
      (**(code **)(plVar8[1] + 0x30))(plVar8);
    }
  }
  local_138 = (longlong *)_PyDict_NewPresized(1);
  PyDict_SetItem(local_138,uVar11,plVar6);
  if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
    (**(code **)(plVar6[1] + 0x30))(plVar6);
  }
  plVar5 = local_128;
  plVar8 = (longlong *)FUN_18037f0c0();
  plVar7 = local_120;
  uVar16 = 0x1f;
  if (plVar8 != (longlong *)0x0) {
    *(undefined4 *)(local_120 + 5) = 0x1f;
    pcVar14 = (code *)FUN_1803770d0(param_1,plVar8,DAT_1803c5500);
    if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
      (**(code **)(plVar8[1] + 0x30))(plVar8);
    }
    if (pcVar14 != (code *)0x0) {
      iVar4 = FUN_180001510(pcVar14);
      if (iVar4 == -1) {
        plVar9 = *(longlong **)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
        plVar8 = local_138;
        if ((-1 < *(int *)pcVar14) &&
           (*(longlong *)pcVar14 = *(longlong *)pcVar14 + -1, *(longlong *)pcVar14 == 0)) {
          (**(code **)(*(longlong *)(pcVar14 + 8) + 0x30))(pcVar14);
          plVar8 = local_138;
        }
        goto LAB_1802d6063;
      }
      if (iVar4 != 1) {
        if ((-1 < *(int *)pcVar14) &&
           (*(longlong *)pcVar14 = *(longlong *)pcVar14 + -1, *(longlong *)pcVar14 == 0)) {
          (**(code **)(*(longlong *)(pcVar14 + 8) + 0x30))(pcVar14);
        }
        pcVar14 = DAT_1803ab250;
        if (*(int *)DAT_1803ab250 + 1 != 0) {
          *(int *)DAT_1803ab250 = *(int *)DAT_1803ab250 + 1;
        }
      }
      plVar6 = (longlong *)PyNumber_Long(pcVar14);
      if ((-1 < *(int *)pcVar14) &&
         (*(longlong *)pcVar14 = *(longlong *)pcVar14 + -1, *(longlong *)pcVar14 == 0)) {
        (**(code **)(*(longlong *)(pcVar14 + 8) + 0x30))(pcVar14);
      }
      if (plVar6 != (longlong *)0x0) {
        local_100 = plVar6;
/* novels.services.state_reducer#12="lastAppliedChapter" */
        pcVar14 = (code *)FUN_18036f6e0(param_1,local_f8,DAT_1803c5540);
        if (pcVar14 == (code *)0x0) {
          pcVar14 = _Py_NoneStruct_exref;
        }
        iVar4 = FUN_180001510(pcVar14);
        if (iVar4 == -1) {
          plVar9 = *(longlong **)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          if ((-1 < *(int *)pcVar14) &&
             (*(longlong *)pcVar14 = *(longlong *)pcVar14 + -1, *(longlong *)pcVar14 == 0)) {
            (**(code **)(*(longlong *)(pcVar14 + 8) + 0x30))(pcVar14);
          }
          uVar16 = 0x20;
          plVar8 = local_138;
          goto LAB_1802d6063;
        }
        if (iVar4 != 1) {
          if ((-1 < *(int *)pcVar14) &&
             (*(longlong *)pcVar14 = *(longlong *)pcVar14 + -1, *(longlong *)pcVar14 == 0)) {
            (**(code **)(*(longlong *)(pcVar14 + 8) + 0x30))(pcVar14);
          }
          pcVar14 = DAT_1803ab250;
          if (*(int *)DAT_1803ab250 + 1 != 0) {
            *(int *)DAT_1803ab250 = *(int *)DAT_1803ab250 + 1;
          }
        }
        plVar8 = (longlong *)PyNumber_Long(pcVar14);
        local_108 = plVar8;
        if ((-1 < *(int *)pcVar14) &&
           (*(longlong *)pcVar14 = *(longlong *)pcVar14 + -1, *(longlong *)pcVar14 == 0)) {
          (**(code **)(*(longlong *)(pcVar14 + 8) + 0x30))(pcVar14);
        }
        if (plVar8 == (longlong *)0x0) {
          uVar16 = 0x20;
        }
        else {
          local_90 = plVar8;
          iVar4 = FUN_180001510(local_c8);
          if (iVar4 == -1) {
            uVar16 = 0x21;
          }
          else if (iVar4 == 0) {
            iVar4 = FUN_180383cc0(plVar6,plVar8);
/* novels.services.state_reducer#13="delta chapter " */
            piVar2 = DAT_1803c5548;
            if (iVar4 != -1) {
              if (iVar4 != 1) goto LAB_1802d5605;
              lVar10 = *(longlong *)(param_1 + 0x10);
              plVar12 = *(longlong **)(lVar10 + 0x41da8);
              if (plVar12 == (longlong *)0x0) {
                plVar12 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,3);
              }
              else {
                *(longlong *)(lVar10 + 0x41da8) = plVar12[3];
                *(int *)(lVar10 + 0x41e40) = *(int *)(lVar10 + 0x41e40) + -1;
                *plVar12 = 1;
              }
              plVar12[3] = 0;
              plVar12[4] = 0;
              plVar12[5] = 0;
              FUN_1801f33c0(plVar12);
              if (*piVar2 + 1 != 0) {
                *piVar2 = *piVar2 + 1;
              }
              plVar12[3] = (longlong)piVar2;
/* novels.services.state_reducer#14="" */
              lVar10 = FUN_180375520(param_1,plVar6,DAT_1803c5550);
              if (lVar10 == 0) {
                plVar9 = *(longlong **)(param_1 + 0x60);
                *(undefined8 *)(param_1 + 0x60) = 0;
                uVar16 = 0x25;
                plVar8 = local_138;
                if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                  (**(code **)(plVar12[1] + 0x30))(plVar12);
                  plVar8 = local_138;
                }
              }
              else {
                plVar12[4] = lVar10;
/* novels.services.state_reducer#15=" goes backwards" */
                piVar2 = DAT_1803c5558;
/* novels.services.state_reducer#15=" goes backwards" */
                if (*DAT_1803c5558 + 1 != 0) {
/* novels.services.state_reducer#15=" goes backwards" */
                  *DAT_1803c5558 = *DAT_1803c5558 + 1;
                }
                plVar12[5] = (longlong)piVar2;
/* novels.services.state_reducer#14="" */
                plVar8 = (longlong *)PyUnicode_Join(DAT_1803c5550,plVar12);
                if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                  (**(code **)(plVar12[1] + 0x30))(plVar12);
                }
                if (plVar8 != (longlong *)0x0) {
                  uVar16 = 0x25;
                  goto LAB_1802d547e;
                }
                plVar9 = *(longlong **)(param_1 + 0x60);
                *(undefined8 *)(param_1 + 0x60) = 0;
                uVar16 = 0x25;
                plVar8 = local_138;
              }
              goto LAB_1802d6063;
            }
            uVar16 = 0x24;
          }
          else {
            iVar4 = FUN_180384640();
/* novels.services.state_reducer#13="delta chapter " */
            piVar2 = DAT_1803c5548;
            if (iVar4 == -1) {
              uVar16 = 0x22;
            }
            else {
              if (iVar4 == 1) {
                lVar10 = *(longlong *)(param_1 + 0x10);
                plVar12 = *(longlong **)(lVar10 + 0x41da8);
                if (plVar12 == (longlong *)0x0) {
                  plVar12 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,3);
                }
                else {
                  *(longlong *)(lVar10 + 0x41da8) = plVar12[3];
                  *(int *)(lVar10 + 0x41e40) = *(int *)(lVar10 + 0x41e40) + -1;
                  *plVar12 = 1;
                }
                plVar12[3] = 0;
                plVar12[4] = 0;
                plVar12[5] = 0;
                FUN_1801f33c0(plVar12);
                if (*piVar2 + 1 != 0) {
                  *piVar2 = *piVar2 + 1;
                }
                plVar12[3] = (longlong)piVar2;
/* novels.services.state_reducer#14="" */
                lVar10 = FUN_180375520(param_1,plVar6,DAT_1803c5550);
                if (lVar10 == 0) {
                  plVar9 = *(longlong **)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                  uVar16 = 0x23;
                  plVar8 = local_138;
                  if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                    (**(code **)(plVar12[1] + 0x30))(plVar12);
                    plVar8 = local_138;
                  }
                }
                else {
                  plVar12[4] = lVar10;
/* novels.services.state_reducer#15=" goes backwards" */
                  piVar2 = DAT_1803c5558;
/* novels.services.state_reducer#15=" goes backwards" */
                  if (*DAT_1803c5558 + 1 != 0) {
/* novels.services.state_reducer#15=" goes backwards" */
                    *DAT_1803c5558 = *DAT_1803c5558 + 1;
                  }
                  plVar12[5] = (longlong)piVar2;
/* novels.services.state_reducer#14="" */
                  plVar8 = (longlong *)PyUnicode_Join(DAT_1803c5550,plVar12);
                  if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                    (**(code **)(plVar12[1] + 0x30))(plVar12);
                  }
                  if (plVar8 == (longlong *)0x0) {
                    plVar9 = *(longlong **)(param_1 + 0x60);
                    *(undefined8 *)(param_1 + 0x60) = 0;
                    uVar16 = 0x23;
                    plVar8 = local_138;
                  }
                  else {
                    uVar16 = 0x23;
LAB_1802d547e:
                    pcVar14 = PyExc_ValueError_exref;
                    *(undefined4 *)(plVar7 + 5) = uVar16;
                    plVar9 = (longlong *)FUN_180376770(param_1,*(undefined8 *)pcVar14,plVar8);
                    if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                      (**(code **)(plVar8[1] + 0x30))(plVar8);
                    }
                    FUN_18000d1d0(param_1,plVar9);
                    plVar8 = local_138;
                  }
                }
                goto LAB_1802d6063;
              }
LAB_1802d5605:
              plVar8 = (longlong *)FUN_18037f0c0();
              uVar16 = 0x27;
              if (plVar8 != (longlong *)0x0) {
                *(undefined4 *)(plVar7 + 5) = 0x27;
                plVar5 = (longlong *)FUN_1803770d0(param_1,plVar8,DAT_1803c5560);
                if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                  (**(code **)(plVar8[1] + 0x30))(plVar8);
                }
                if (plVar5 == (longlong *)0x0) goto LAB_1802d565c;
                local_88 = plVar5;
                iVar4 = PyObject_IsInstance(plVar5,PyDict_Type_exref);
                plVar8 = local_138;
                if (iVar4 == -1) {
                  plVar9 = *(longlong **)(param_1 + 0x60);
                  uVar16 = 0x28;
                  *(undefined8 *)(param_1 + 0x60) = 0;
                  plVar5 = local_128;
                  goto LAB_1802d6063;
                }
                if (iVar4 != 0) {
                  plVar8 = (longlong *)FUN_18037f0c0();
                  uVar16 = 0x29;
                  if (plVar8 != (longlong *)0x0) {
                    *(undefined4 *)(plVar7 + 5) = 0x29;
                    pcVar14 = (code *)FUN_1803770d0(param_1,plVar8,DAT_1803c5500);
                    if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                      (**(code **)(plVar8[1] + 0x30))(plVar8);
                    }
                    if (pcVar14 != (code *)0x0) {
                      iVar4 = FUN_180001510(pcVar14);
                      if (iVar4 == -1) {
                        plVar9 = *(longlong **)(param_1 + 0x60);
                        *(undefined8 *)(param_1 + 0x60) = 0;
                        plVar5 = local_128;
                        plVar8 = local_138;
                        if ((-1 < *(int *)pcVar14) &&
                           (*(longlong *)pcVar14 = *(longlong *)pcVar14 + -1,
                           *(longlong *)pcVar14 == 0)) {
                          (**(code **)(*(longlong *)(pcVar14 + 8) + 0x30))(pcVar14);
                          plVar5 = local_128;
                          plVar8 = local_138;
                        }
                        goto LAB_1802d6063;
                      }
                      if (iVar4 != 1) {
                        if ((-1 < *(int *)pcVar14) &&
                           (*(longlong *)pcVar14 = *(longlong *)pcVar14 + -1,
                           *(longlong *)pcVar14 == 0)) {
                          (**(code **)(*(longlong *)(pcVar14 + 8) + 0x30))(pcVar14);
                        }
                        pcVar14 = DAT_1803ab250;
                        if (*(int *)DAT_1803ab250 + 1 != 0) {
                          *(int *)DAT_1803ab250 = *(int *)DAT_1803ab250 + 1;
                        }
                      }
                      lVar10 = PyNumber_Long(pcVar14);
                      if ((-1 < *(int *)pcVar14) &&
                         (*(longlong *)pcVar14 = *(longlong *)pcVar14 + -1,
                         *(longlong *)pcVar14 == 0)) {
                        (**(code **)(*(longlong *)(pcVar14 + 8) + 0x30))(pcVar14);
                      }
                      plVar9 = local_a8;
                      if (lVar10 != 0) {
                        local_a8[2] = lVar10;
                        iVar4 = FUN_180385f30(lVar10,plVar6);
                        plVar8 = local_138;
/* novels.services.state_reducer#17="chapter summary " */
                        piVar2 = DAT_1803c5568;
                        if (iVar4 != -1) {
                          if (iVar4 == 1) {
                            lVar10 = *(longlong *)(param_1 + 0x10);
                            plVar12 = *(longlong **)(lVar10 + 0x41db0);
                            if (plVar12 == (longlong *)0x0) {
                              plVar12 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,4);
                            }
                            else {
                              *(longlong *)(lVar10 + 0x41db0) = plVar12[3];
                              *(int *)(lVar10 + 0x41e44) = *(int *)(lVar10 + 0x41e44) + -1;
                              *plVar12 = 1;
                            }
                            plVar12[3] = 0;
                            plVar12[4] = 0;
                            plVar12[5] = 0;
                            plVar12[6] = 0;
                            FUN_1801f33c0(plVar12);
                            if (*piVar2 + 1 != 0) {
                              *piVar2 = *piVar2 + 1;
                            }
                            plVar12[3] = (longlong)piVar2;
/* novels.services.state_reducer#14="" */
                            lVar10 = FUN_180375520(param_1,local_a8[2],DAT_1803c5550);
                            if (lVar10 != 0) {
                              plVar12[4] = lVar10;
/* novels.services.state_reducer#18=" does not match delta chapter " */
                              piVar2 = DAT_1803c5570;
/* novels.services.state_reducer#18=" does not match delta chapter " */
                              if (*DAT_1803c5570 + 1 != 0) {
/* novels.services.state_reducer#18=" does not match delta chapter " */
                                *DAT_1803c5570 = *DAT_1803c5570 + 1;
                              }
                              plVar12[5] = (longlong)piVar2;
/* novels.services.state_reducer#14="" */
                              lVar10 = FUN_180375520(param_1,plVar6,DAT_1803c5550);
                              if (lVar10 != 0) {
                                plVar12[6] = lVar10;
/* novels.services.state_reducer#14="" */
                                plVar8 = (longlong *)PyUnicode_Join(DAT_1803c5550,plVar12);
                                if ((-1 < (int)*plVar12) &&
                                   (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                  (**(code **)(plVar12[1] + 0x30))(plVar12);
                                }
                                pcVar14 = PyExc_ValueError_exref;
                                if (plVar8 == (longlong *)0x0) {
                                  plVar9 = *(longlong **)(param_1 + 0x60);
                                  uVar16 = 0x2c;
                                  *(undefined8 *)(param_1 + 0x60) = 0;
                                  plVar5 = local_128;
                                  plVar8 = local_138;
                                  goto LAB_1802d6063;
                                }
                                uVar16 = 0x2b;
                                *(undefined4 *)(plVar7 + 5) = 0x2b;
                                plVar9 = (longlong *)
                                         FUN_180376770(param_1,*(undefined8 *)pcVar14,plVar8);
                                iVar4 = (int)*plVar8;
LAB_1802d5931:
                                if ((-1 < iVar4) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                  (**(code **)(plVar8[1] + 0x30))(plVar8);
                                }
                                FUN_18000d1d0(param_1,plVar9);
                                plVar5 = local_128;
                                plVar8 = local_138;
                                goto LAB_1802d6063;
                              }
                            }
                            uVar16 = 0x2c;
                            plVar9 = *(longlong **)(param_1 + 0x60);
                            *(undefined8 *)(param_1 + 0x60) = 0;
                            plVar5 = local_128;
                            plVar8 = local_138;
                            if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                              (**(code **)(plVar12[1] + 0x30))(plVar12);
                              plVar5 = local_128;
                              plVar8 = local_138;
                            }
                          }
                          else {
/* novels.services.state_reducer#9="rows" */
                            plVar6 = (longlong *)FUN_18036f800(param_1,local_138,DAT_1803c5528);
                            if (plVar6 != (longlong *)0x0) {
                              plVar12 = (longlong *)FUN_180001930();
                              if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                (**(code **)(plVar6[1] + 0x30))(plVar6);
                              }
                              if (plVar12 != (longlong *)0x0) {
                                local_118 = (longlong *)FUN_18036d750(plVar12);
                                local_110 = plVar9;
                                iVar4 = (int)*plVar9 + 1;
                                if (iVar4 != 0) {
                                  *(int *)plVar9 = iVar4;
                                }
                                local_180 = &local_118;
                                local_170 = (longlong *)0x448;
                                local_178 = (longlong *)0x2;
                                local_188 = (longlong *)DAT_1803c5400;
                                plVar6 = (longlong *)
/* novels.services.state_reducer#27="<genexpr>" */
                                         FUN_180367890(FUN_1802d6360,DAT_1803ca450,DAT_1803c55b8,
/* novels.services.state_reducer#28="apply_runtime_state_delta.<locals>.<genexpr>" */
                                                       DAT_1803c55c0);
                                if ((-1 < (int)*plVar12) &&
                                   (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                  (**(code **)(plVar12[1] + 0x30))(plVar12);
                                }
                                plVar9 = (longlong *)FUN_1803753b0(param_1,plVar6);
                                if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                  (**(code **)(plVar6[1] + 0x30))(plVar6);
                                }
                                if (plVar9 != (longlong *)0x0) {
                                  iVar4 = FUN_180001510(plVar9);
                                  bVar17 = iVar4 == 1;
                                  if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0))
                                  {
                                    (**(code **)(plVar9[1] + 0x30))(plVar9);
                                  }
                                  if (iVar4 == 1) {
                                    iVar4 = FUN_180001510(local_c8);
                                    if (iVar4 == -1) goto LAB_1802d597d;
                                    bVar17 = iVar4 == 0;
                                  }
/* novels.services.state_reducer#19="duplicate summary row for chapter " */
                                  piVar2 = DAT_1803c5578;
                                  plVar6 = local_100;
                                  if (!bVar17) goto LAB_1802d5be8;
                                  lVar10 = *(longlong *)(param_1 + 0x10);
                                  plVar6 = *(longlong **)(lVar10 + 0x41da0);
                                  if (plVar6 == (longlong *)0x0) {
                                    plVar6 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,2);
                                  }
                                  else {
                                    *(longlong *)(lVar10 + 0x41da0) = plVar6[3];
                                    *(int *)(lVar10 + 0x41e3c) = *(int *)(lVar10 + 0x41e3c) + -1;
                                    *plVar6 = 1;
                                  }
                                  plVar6[3] = 0;
                                  plVar6[4] = 0;
                                  FUN_1801f33c0(plVar6);
                                  if (*piVar2 + 1 != 0) {
                                    *piVar2 = *piVar2 + 1;
                                  }
                                  plVar6[3] = (longlong)piVar2;
/* novels.services.state_reducer#14="" */
                                  lVar10 = FUN_180375520(param_1,local_a8[2],DAT_1803c5550);
                                  if (lVar10 == 0) {
                                    plVar9 = *(longlong **)(param_1 + 0x60);
                                    *(undefined8 *)(param_1 + 0x60) = 0;
                                    uVar16 = 0x2f;
                                    plVar5 = local_128;
                                    if ((-1 < (int)*plVar6) &&
                                       (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                      (**(code **)(plVar6[1] + 0x30))(plVar6);
                                      plVar5 = local_128;
                                    }
                                  }
                                  else {
                                    plVar6[4] = lVar10;
/* novels.services.state_reducer#14="" */
                                    plVar8 = (longlong *)PyUnicode_Join(DAT_1803c5550,plVar6);
                                    if ((-1 < (int)*plVar6) &&
                                       (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                      (**(code **)(plVar6[1] + 0x30))(plVar6);
                                    }
                                    pcVar14 = PyExc_ValueError_exref;
                                    if (plVar8 != (longlong *)0x0) {
                                      uVar16 = 0x2f;
                                      *(undefined4 *)(plVar7 + 5) = 0x2f;
                                      plVar9 = (longlong *)
                                               FUN_180376770(param_1,*(undefined8 *)pcVar14,plVar8);
                                      iVar4 = (int)*plVar8;
                                      goto LAB_1802d5931;
                                    }
                                    plVar9 = *(longlong **)(param_1 + 0x60);
                                    *(undefined8 *)(param_1 + 0x60) = 0;
                                    uVar16 = 0x2f;
                                    plVar5 = local_128;
                                    plVar8 = local_138;
                                  }
                                  goto LAB_1802d6063;
                                }
                              }
                            }
LAB_1802d597d:
                            plVar9 = *(longlong **)(param_1 + 0x60);
                            uVar16 = 0x2e;
                            *(undefined8 *)(param_1 + 0x60) = 0;
                            plVar5 = local_128;
                          }
                          goto LAB_1802d6063;
                        }
                        uVar16 = 0x2a;
                      }
                    }
                  }
                  goto LAB_1802d565c;
                }
LAB_1802d5be8:
                lVar10 = FUN_1802d3200(param_1);
                plVar9 = local_b0;
                if (lVar10 == 0) {
/* novels.services.state_reducer#20="_apply_hook_ops" */
                  FUN_180373340(param_1,&local_d0,DAT_1803c5580);
                  uVar16 = 0x31;
                  plVar9 = local_d0;
                  plVar5 = local_128;
                }
                else {
                  uVar16 = 0x31;
                  local_68 = local_128;
                  *(undefined4 *)(plVar7 + 5) = 0x31;
                  local_70 = local_b0;
                  local_60 = plVar6;
                  plVar6 = (longlong *)FUN_180378890(param_1,lVar10,&local_70);
                  if (plVar6 == (longlong *)0x0) {
                    plVar9 = *(longlong **)(param_1 + 0x60);
                    *(undefined8 *)(param_1 + 0x60) = 0;
                    plVar7 = local_120;
                    plVar5 = local_128;
                  }
                  else {
                    local_b0 = plVar6;
                    if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                      (**(code **)(plVar9[1] + 0x30))(plVar9);
                    }
                    lVar10 = FUN_1802d30f0(param_1);
                    if (lVar10 == 0) {
/* novels.services.state_reducer#21="_apply_current_state_patch" */
                      FUN_180373340(param_1,&local_d0,DAT_1803c5588);
                      uVar16 = 0x32;
                      plVar9 = local_d0;
                      plVar7 = local_120;
                      plVar5 = local_128;
                    }
                    else {
/* novels.services.state_reducer#22="language" */
                      pcVar14 = (code *)FUN_18036f6e0(param_1,local_f8,DAT_1803c5590);
                      if (pcVar14 == (code *)0x0) {
                        pcVar14 = _Py_NoneStruct_exref;
                      }
                      iVar4 = FUN_180001510(pcVar14);
                      if (iVar4 == -1) {
                        plVar9 = *(longlong **)(param_1 + 0x60);
                        *(undefined8 *)(param_1 + 0x60) = 0;
                        if ((-1 < *(int *)pcVar14) &&
                           (*(longlong *)pcVar14 = *(longlong *)pcVar14 + -1,
                           *(longlong *)pcVar14 == 0)) {
                          (**(code **)(*(longlong *)(pcVar14 + 8) + 0x30))(pcVar14);
                        }
                        uVar16 = 0x34;
                        plVar7 = local_120;
                        plVar5 = local_128;
                      }
                      else {
                        if (iVar4 != 1) {
                          if ((-1 < *(int *)pcVar14) &&
                             (*(longlong *)pcVar14 = *(longlong *)pcVar14 + -1,
                             *(longlong *)pcVar14 == 0)) {
                            (**(code **)(*(longlong *)(pcVar14 + 8) + 0x30))(pcVar14);
                          }
/* novels.services.state_reducer#23="zh" */
                          pcVar14 = DAT_1803c5598;
/* novels.services.state_reducer#23="zh" */
                          if (*(int *)DAT_1803c5598 + 1 != 0) {
/* novels.services.state_reducer#23="zh" */
                            *(int *)DAT_1803c5598 = *(int *)DAT_1803c5598 + 1;
                          }
                        }
                        plVar7 = (longlong *)FUN_1803749a0(pcVar14);
                        if ((-1 < *(int *)pcVar14) &&
                           (*(longlong *)pcVar14 = *(longlong *)pcVar14 + -1,
                           *(longlong *)pcVar14 == 0)) {
                          (**(code **)(*(longlong *)(pcVar14 + 8) + 0x30))(pcVar14);
                        }
                        plVar9 = local_80;
                        if (plVar7 == (longlong *)0x0) {
                          plVar9 = *(longlong **)(param_1 + 0x60);
                          uVar16 = 0x34;
                          *(undefined8 *)(param_1 + 0x60) = 0;
                          plVar7 = local_120;
                          plVar5 = local_128;
                        }
                        else {
                          uVar16 = 0x32;
                          local_58 = local_80;
                          *(undefined4 *)(local_120 + 5) = 0x32;
                          local_48 = local_128;
                          local_50 = plVar7;
                          plVar8 = (longlong *)FUN_180378890(param_1,lVar10,&local_58);
                          if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                            (**(code **)(plVar7[1] + 0x30))(plVar7);
                          }
                          plVar7 = local_120;
                          if (plVar8 != (longlong *)0x0) {
                            local_98 = plVar8;
                            if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                              (**(code **)(plVar9[1] + 0x30))(plVar9);
                            }
                            lVar10 = FUN_1802d3310(param_1);
                            plVar9 = local_138;
                            if (lVar10 == 0) {
/* novels.services.state_reducer#24="_apply_summary_delta" */
                              FUN_180373340(param_1,&local_d0,DAT_1803c55a0);
                              uVar16 = 0x37;
                              plVar9 = local_d0;
                              plVar7 = local_120;
                              plVar5 = local_128;
                              plVar8 = local_138;
                              goto LAB_1802d6063;
                            }
                            uVar16 = 0x37;
                            local_80 = local_138;
                            *(undefined4 *)(local_120 + 5) = 0x37;
                            local_78 = local_128;
                            local_118 = local_c8;
                            local_188 = (longlong *)DAT_1803c55a8;
                            plVar12 = (longlong *)FUN_18037d650(param_1,lVar10,&local_80,&local_118)
                            ;
                            plVar7 = local_120;
                            if (plVar12 != (longlong *)0x0) {
                              if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                (**(code **)(plVar9[1] + 0x30))(plVar9);
                              }
                              FUN_180001d80(param_1);
                              plVar7 = local_100;
/* novels.services.state_reducer#12="lastAppliedChapter" */
                              PyDict_SetItem(local_f8,DAT_1803c5540,local_100);
/* novels.services.state_reducer#0="manifest" */
                              uVar11 = DAT_1803c54e0;
                              uVar13 = _PyDict_NewPresized(4);
                              PyDict_SetItem(uVar13,uVar11,local_f8);
/* novels.services.state_reducer#2="currentState" */
                              PyDict_SetItem(uVar13,DAT_1803c54f0,plVar8);
/* novels.services.state_reducer#7="hooks" */
                              PyDict_SetItem(uVar13,DAT_1803c5518,plVar6);
/* novels.services.state_reducer#10="chapterSummaries" */
                              PyDict_SetItem(uVar13,DAT_1803c5530,plVar12);
                              if ((-1 < (int)*local_130) &&
                                 (*local_130 = *local_130 + -1, *local_130 == 0)) {
                                (**(code **)(local_f8[1] + 0x30))();
                              }
                              if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                (**(code **)(plVar8[1] + 0x30))(plVar8);
                              }
                              if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                (**(code **)(plVar6[1] + 0x30))(plVar6);
                              }
                              if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0))
                              {
                                (**(code **)(plVar12[1] + 0x30))(plVar12);
                              }
                              if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                (**(code **)(plVar7[1] + 0x30))(plVar7);
                              }
                              if ((-1 < (int)*local_108) &&
                                 (*local_108 = *local_108 + -1, *local_108 == 0)) {
                                (**(code **)(local_90[1] + 0x30))();
                              }
                              if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                (**(code **)(plVar5[1] + 0x30))(plVar5);
                              }
                              if ((-1 < (int)*local_a8) &&
                                 (*local_a8 = *local_a8 + -1, *local_a8 == 0)) {
                                (**(code **)(local_a8[1] + 0x30))();
                              }
                              if ((-1 < (int)*local_a0) &&
                                 (*local_a0 = *local_a0 + -1, *local_a0 == 0)) {
                                (**(code **)(local_a0[1] + 0x30))();
                              }
                              if ((-1 < (int)*local_128) &&
                                 (*local_128 = *local_128 + -1, *local_128 == 0)) {
                                (**(code **)(local_128[1] + 0x30))();
                              }
                              if (-1 < (int)*local_c8) {
                                *local_c8 = *local_c8 + -1;
                                if (*local_c8 == 0) {
                                  (**(code **)(local_c8[1] + 0x30))();
                                  return uVar13;
                                }
                                return uVar13;
                              }
                              return uVar13;
                            }
                          }
LAB_1802d565c:
                          plVar9 = *(longlong **)(param_1 + 0x60);
                          *(undefined8 *)(param_1 + 0x60) = 0;
                          plVar5 = local_128;
                          plVar8 = local_138;
                        }
                      }
                    }
                  }
                }
                goto LAB_1802d6063;
              }
            }
          }
        }
      }
    }
  }
  plVar9 = *(longlong **)(param_1 + 0x60);
  *(undefined8 *)(param_1 + 0x60) = 0;
  plVar8 = local_138;
  goto LAB_1802d6063;
  while( true ) {
    plVar9 = (longlong *)(*pcVar14)(plVar8);
    if (plVar9 == (longlong *)0x0) goto LAB_1802d4b6c;
    local_b8 = plVar9;
    if (((plVar6 != (longlong *)0x0) && (-1 < (int)*plVar6)) &&
       (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
      (**(code **)(plVar6[1] + 0x30))(plVar6);
    }
    iVar4 = (int)*plVar9 + 1;
    if (iVar4 != 0) {
      *(int *)plVar9 = iVar4;
    }
    local_e8 = plVar9;
    if (((plVar5 != (longlong *)0x0) && (-1 < (int)*plVar5)) &&
       (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
      (**(code **)(plVar5[1] + 0x30))();
    }
    iVar4 = PyObject_IsInstance(plVar9,PyDict_Type_exref);
    if (iVar4 == -1) goto LAB_1802d4b46;
    if (iVar4 != 0) {
      lVar10 = FUN_180370dc0(param_1,plVar9);
      if (lVar10 == 0) goto LAB_1802d4b46;
      lVar1 = plVar7[2];
      cVar3 = FUN_1803722b0(plVar7,lVar1 + 1);
      if (cVar3 == '\0') goto LAB_1802d4b46;
      *(longlong *)(plVar7[3] + lVar1 * 8) = lVar10;
    }
    lVar10 = *(longlong *)(param_1 + 0x10);
    if (((*(int *)(*(longlong *)(lVar10 + 0x60) + 0x4d4) != 0) || (*(int *)(lVar10 + 0x1a0) != 0))
       && ((iVar4 = Py_MakePendingCalls(), iVar4 < 0 &&
           (plVar9 = *(longlong **)(param_1 + 0x60), plVar9 != (longlong *)0x0))))
    goto LAB_1802d4bc4;
    if (*(int *)(lVar10 + 0x174) != 0) {
      PyEval_SaveThread();
      PyEval_AcquireThread(param_1);
    }
    if (*(longlong *)(param_1 + 0x80) != 0) break;
LAB_1802d49d0:
    plVar6 = local_b8;
    plVar5 = local_e8;
    pcVar14 = *(code **)(plVar8[1] + 0xe0);
    if (pcVar14 == (code *)0x0) {
      PyErr_Format(*(undefined8 *)PyExc_TypeError_exref,"\'%s\' object is not an iterator",
                   *(undefined8 *)(plVar8[1] + 0x18));
LAB_1802d4b6c:
      plVar9 = *(longlong **)(param_1 + 0x60);
      if (plVar9 != (longlong *)0x0) {
        plVar12 = (longlong *)plVar9[1];
        if ((*(uint *)((longlong *)plVar9[1] + 0x15) & 0x40000000) == 0) {
          plVar12 = plVar9;
        }
        if ((plVar12 != *(longlong **)PyExc_StopIteration_exref) &&
           (((-1 < *(int *)(plVar12[1] + 0xa8) || ((*(uint *)(plVar12 + 0x15) & 0x40000000) == 0))
            || (cVar3 = FUN_180382640(), cVar3 != '\x01')))) goto LAB_1802d4bc4;
        *(undefined8 *)(param_1 + 0x60) = 0;
        if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
          (**(code **)(plVar9[1] + 0x30))(plVar9);
        }
      }
      iVar4 = (int)*plVar7 + 1;
      if (iVar4 != 0) {
        *(int *)plVar7 = iVar4;
      }
      if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
        (**(code **)(plVar8[1] + 0x30))(plVar8);
      }
      if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
        (**(code **)(plVar7[1] + 0x30))(plVar7);
      }
      if (((plVar6 != (longlong *)0x0) && (-1 < (int)*plVar6)) &&
         (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
        (**(code **)(plVar6[1] + 0x30))(plVar6);
      }
      if (((plVar5 != (longlong *)0x0) && (-1 < (int)*plVar5)) &&
         (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
        (**(code **)(plVar5[1] + 0x30))(plVar5);
      }
      plVar8 = (longlong *)_PyDict_NewPresized(1);
      PyDict_SetItem(plVar8,local_f0,plVar7);
      if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
        (**(code **)(plVar7[1] + 0x30))(plVar7);
      }
/* novels.services.state_reducer#9="rows" */
      uVar11 = DAT_1803c5528;
      local_b0 = plVar8;
/* novels.services.state_reducer#10="chapterSummaries" */
      plVar8 = (longlong *)FUN_180024e50(param_1,local_a0,DAT_1803c5530);
      uVar16 = 0x1c;
      if (plVar8 == (longlong *)0x0) goto LAB_1802d4289;
      plVar5 = (longlong *)FUN_18037f0c0();
      if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
        (**(code **)(plVar8[1] + 0x30))(plVar8);
      }
      plVar7 = local_120;
      if (plVar5 == (longlong *)0x0) {
        plVar9 = *(longlong **)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
        plVar5 = local_128;
        plVar8 = local_138;
        goto LAB_1802d6063;
      }
      *(undefined4 *)(local_120 + 5) = 0x1c;
      plVar6 = (longlong *)FUN_1803770d0(param_1,plVar5,DAT_1803c5538);
      if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
        (**(code **)(plVar5[1] + 0x30))(plVar5);
      }
      if (plVar6 == (longlong *)0x0) {
        plVar9 = *(longlong **)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
        plVar5 = local_128;
        plVar8 = local_138;
        goto LAB_1802d6063;
      }
      iVar4 = FUN_180001510(plVar6);
      if (iVar4 == -1) goto LAB_1802d4925;
      if (iVar4 != 1) {
        if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
          (**(code **)(plVar6[1] + 0x30))(plVar6);
        }
        plVar6 = (longlong *)FUN_180372160(param_1,0);
      }
      plVar8 = (longlong *)FUN_180001930();
      if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
        (**(code **)(plVar6[1] + 0x30))(plVar6);
      }
      if (plVar8 == (longlong *)0x0) goto LAB_1802d4289;
      plVar6 = (longlong *)FUN_180372160(param_1,0);
      plVar7 = (longlong *)FUN_18000d430(plVar8);
      plVar5 = local_d8;
      goto joined_r0x0001802d4e75;
    }
  }
  plVar5 = *(longlong **)(param_1 + 0x60);
  *(undefined8 *)(param_1 + 0x80) = 0;
  plVar9 = (longlong *)FUN_180373490(param_1);
  *(longlong **)(param_1 + 0x60) = plVar9;
  if ((plVar5 != (longlong *)0x0) && (-1 < (int)*plVar5)) {
    *plVar5 = *plVar5 + -1;
    if (*plVar5 == 0) {
      (**(code **)(plVar5[1] + 0x30))(plVar5);
    }
LAB_1802d4b46:
    plVar9 = *(longlong **)(param_1 + 0x60);
  }
LAB_1802d4bc4:
  *(undefined8 *)(param_1 + 0x60) = 0;
  if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
    (**(code **)(plVar8[1] + 0x30))(plVar8);
  }
  if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
    (**(code **)(plVar7[1] + 0x30))(plVar7);
  }
  plVar6 = local_e8;
  if (((local_b8 != (longlong *)0x0) && (-1 < (int)*local_b8)) &&
     (*local_b8 = *local_b8 + -1, *local_b8 == 0)) {
    (**(code **)(local_b8[1] + 0x30))();
    plVar6 = local_e8;
  }
LAB_1802d4c25:
  plVar7 = local_120;
  plVar5 = local_128;
  plVar8 = local_138;
  if (plVar6 != (longlong *)0x0) {
LAB_1802d4c2e:
    plVar7 = local_120;
    plVar5 = local_128;
    plVar8 = local_138;
    if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
      (**(code **)(plVar6[1] + 0x30))();
      plVar7 = local_120;
      plVar5 = local_128;
      plVar8 = local_138;
    }
  }
LAB_1802d6063:
  piVar2 = (int *)plVar9[5];
  if (piVar2 == (int *)0x0) {
    pcVar14 = (code *)FUN_18038e2a0(plVar7,uVar16);
    plVar6 = (longlong *)plVar9[5];
    pcVar15 = pcVar14;
    if (pcVar14 == _Py_NoneStruct_exref) {
      pcVar15 = (code *)0x0;
    }
    else {
LAB_1802d60bd:
      pcVar14 = pcVar15;
      if ((pcVar15 != (code *)0x0) && (*(int *)pcVar15 + 1 != 0)) {
        *(int *)pcVar15 = *(int *)pcVar15 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar2 + 6) == plVar7) goto LAB_1802d6106;
    pcVar14 = (code *)FUN_18038e2a0(plVar7,uVar16);
    *(int **)(pcVar14 + 0x10) = piVar2;
    if (*piVar2 + 1 != 0) {
      *piVar2 = *piVar2 + 1;
    }
    plVar6 = (longlong *)plVar9[5];
    pcVar15 = pcVar14;
    if (pcVar14 != _Py_NoneStruct_exref) goto LAB_1802d60bd;
    pcVar15 = (code *)0x0;
  }
  plVar9[5] = (longlong)pcVar15;
  if (((plVar6 != (longlong *)0x0) && (-1 < (int)*plVar6)) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)
     ) {
    (**(code **)(plVar6[1] + 0x30))();
  }
  if (((pcVar14 != (code *)0x0) && (-1 < *(int *)pcVar14)) &&
     (*(longlong *)pcVar14 = *(longlong *)pcVar14 + -1, *(longlong *)pcVar14 == 0)) {
    (**(code **)(*(longlong *)(pcVar14 + 8) + 0x30))(pcVar14);
  }
LAB_1802d6106:
  plVar12 = local_a8;
  local_148 = local_a8;
  local_150 = local_88;
  local_158 = local_90;
  local_160 = local_100;
  local_170 = local_b0;
  local_178 = local_98;
  local_180 = (longlong **)local_f8;
  local_188 = local_c8;
  local_168 = plVar8;
  FUN_18036d0a0(plVar7,"ooooooooooc",local_a0,plVar5);
  plVar6 = DAT_1803c53f8;
  if (plVar7 == DAT_1803c53f8) {
    if ((-1 < (int)*DAT_1803c53f8) && (*DAT_1803c53f8 = *DAT_1803c53f8 + -1, *plVar6 == 0)) {
      (**(code **)(plVar6[1] + 0x30))();
    }
    DAT_1803c53f8 = (longlong *)0x0;
  }
  lVar10 = **(longlong **)(param_1 + 0x38);
  plVar7 = *(longlong **)(lVar10 + 0x30);
  **(longlong **)(param_1 + 0x38) = *(longlong *)(lVar10 + 8);
  *(undefined4 *)(plVar7 + 8) = 0xffffffff;
  plVar6 = (longlong *)plVar7[2];
  if (((plVar6 != (longlong *)0x0) && (plVar7[2] = 0, -1 < (int)*plVar6)) &&
     (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
    (**(code **)(plVar6[1] + 0x30))();
  }
  if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
    (**(code **)(plVar7[1] + 0x30))(plVar7);
  }
  plVar7[10] = 0;
  if (((local_f8 != (longlong *)0x0) && (-1 < (int)*local_f8)) &&
     (*local_f8 = *local_f8 + -1, *local_f8 == 0)) {
    (**(code **)(local_f8[1] + 0x30))();
  }
  if (((local_98 != (longlong *)0x0) && (-1 < (int)*local_98)) &&
     (*local_98 = *local_98 + -1, *local_98 == 0)) {
    (**(code **)(local_98[1] + 0x30))();
  }
  if (((local_b0 != (longlong *)0x0) && (-1 < (int)*local_b0)) &&
     (*local_b0 = *local_b0 + -1, *local_b0 == 0)) {
    (**(code **)(local_b0[1] + 0x30))();
  }
  if (((plVar8 != (longlong *)0x0) && (-1 < (int)*plVar8)) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)
     ) {
    (**(code **)(plVar8[1] + 0x30))(plVar8);
  }
  if (((local_100 != (longlong *)0x0) && (-1 < (int)*local_100)) &&
     (*local_100 = *local_100 + -1, *local_100 == 0)) {
    (**(code **)(local_100[1] + 0x30))();
  }
  if (((local_90 != (longlong *)0x0) && (-1 < (int)*local_90)) &&
     (*local_90 = *local_90 + -1, *local_90 == 0)) {
    (**(code **)(local_90[1] + 0x30))();
  }
  if (((local_88 != (longlong *)0x0) && (-1 < (int)*local_88)) &&
     (*local_88 = *local_88 + -1, *local_88 == 0)) {
    (**(code **)(local_88[1] + 0x30))();
  }
  if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
    (**(code **)(plVar12[1] + 0x30))(plVar12);
  }
  if ((-1 < (int)*local_a0) && (*local_a0 = *local_a0 + -1, *local_a0 == 0)) {
    (**(code **)(local_a0[1] + 0x30))();
  }
  if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
    (**(code **)(plVar5[1] + 0x30))(plVar5);
  }
  if ((-1 < (int)*local_c8) && (*local_c8 = *local_c8 + -1, *local_c8 == 0)) {
    (**(code **)(local_c8[1] + 0x30))();
  }
  plVar8 = *(longlong **)(param_1 + 0x60);
  *(longlong **)(param_1 + 0x60) = plVar9;
  if (((plVar8 != (longlong *)0x0) && (-1 < (int)*plVar8)) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)
     ) {
    (**(code **)(plVar8[1] + 0x30))();
  }
  return 0;
}
