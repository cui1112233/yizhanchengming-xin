
int * FUN_1800c10e0(longlong param_1,int *param_2,undefined8 param_3)

{
  int *piVar1;
  undefined8 uVar2;
  undefined8 uVar3;
  int *piVar4;
  char cVar5;
  undefined8 *puVar6;
  longlong *plVar7;
  undefined8 uVar8;
  longlong lVar9;
  longlong lVar10;
  longlong *plVar11;
  longlong *plVar12;
  longlong *plVar13;
  longlong lVar14;
  code *pcVar15;
  longlong *plVar16;
  undefined8 uVar17;
  code *pcVar18;
  int iVar19;
  code *pcVar20;
  undefined8 uVar21;
  bool bVar22;
  longlong *local_res20;
  undefined8 in_stack_ffffffffffffff68;
  undefined4 uVar23;
  
  uVar23 = (undefined4)((ulonglong)in_stack_ffffffffffffff68 >> 0x20);
  DAT_1803ca7b8 = *(undefined8 *)(param_2 + 4);
  DAT_1803ca7c0 = param_2;
  if (DAT_1803adf9f == '\0') {
    if (DAT_1803adf9e == '\0') {
      FUN_180388a30(param_1,&DAT_1803b1f60,"novels.pipeline.chapter_review_cycle");
      DAT_1803adf9e = '\x01';
    }
    FUN_1800b5340();
    DAT_1803adf9f = '\x01';
  }
  uVar17 = DAT_1803ca7b8;
  plVar7 = DAT_1803ab560;
  uVar8 = DAT_1803ab330;
  puVar6 = (undefined8 *)FUN_1800015a0(DAT_1803ca7b8,DAT_1803ab330);
  if (puVar6 == (undefined8 *)0x0) {
LAB_1800c119e:
    PyDict_SetItem(uVar17,uVar8,plVar7);
  }
  else {
    plVar12 = (longlong *)*puVar6;
    if (plVar7 != plVar12) {
      if (plVar12 == (longlong *)0x0) goto LAB_1800c119e;
      iVar19 = (int)*plVar7 + 1;
      if (iVar19 != 0) {
        *(int *)plVar7 = iVar19;
      }
      *puVar6 = plVar7;
      if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
        (**(code **)(plVar12[1] + 0x30))(plVar12);
      }
    }
  }
  plVar7 = (longlong *)FUN_1800015a0(DAT_1803ca7b8,DAT_1803ab2a8);
  pcVar20 = (code *)0x0;
  pcVar18 = pcVar20;
  if (plVar7 != (longlong *)0x0) {
    pcVar18 = (code *)*plVar7;
  }
  uVar8 = PyUnicode_GetLength(pcVar18);
  lVar9 = PyUnicode_Find(pcVar18,DAT_1803ab4a8,0,uVar8,CONCAT44(uVar23,0xffffffff));
  if (lVar9 != -1) {
    PyUnicode_Substring(pcVar18,0,lVar9);
    FUN_1800b47a0();
  }
  plVar7 = (longlong *)FUN_1800015a0(DAT_1803ca7b8,DAT_1803ab300);
  if ((plVar7 == (longlong *)0x0) || (*plVar7 == 0)) {
    PyModule_GetDict(DAT_1803ab780);
    FUN_1800b4710();
  }
  plVar7 = (longlong *)FUN_18038d790(param_3);
  uVar17 = DAT_1803ca7b8;
  uVar8 = DAT_1803ab328;
  puVar6 = (undefined8 *)FUN_1800015a0(DAT_1803ca7b8,DAT_1803ab328);
  if (puVar6 == (undefined8 *)0x0) {
LAB_1800c12ab:
    PyDict_SetItem(uVar17,uVar8,plVar7);
  }
  else {
    plVar12 = (longlong *)*puVar6;
    if (plVar7 != plVar12) {
      if (plVar12 == (longlong *)0x0) goto LAB_1800c12ab;
      iVar19 = (int)*plVar7 + 1;
      if (iVar19 != 0) {
        *(int *)plVar7 = iVar19;
      }
      *puVar6 = plVar7;
      if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
        (**(code **)(plVar12[1] + 0x30))(plVar12);
      }
    }
  }
  if (DAT_1803c9fd0 == 0) {
    DAT_1803c9fd0 = PyImport_ImportModule("importlib._bootstrap");
  }
  plVar7 = (longlong *)PyObject_GetAttrString(DAT_1803c9fd0,"_spec_from_module");
  lVar9 = FUN_180376770(param_1,plVar7,DAT_1803ca7c0);
  if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
    (**(code **)(plVar7[1] + 0x30))(plVar7);
  }
  if (lVar9 == 0) {
    PyErr_PrintEx(0);
                    /* WARNING: Subroutine does not return */
    abort();
  }
  FUN_18037f3a0();
  FUN_1800b47a0();
  local_res20 = (longlong *)0x0;
  FUN_1800b4710();
  FUN_1800b4710();
  piVar4 = DAT_1803ca7c0;
  piVar1 = *(int **)(DAT_1803ca7c0 + 4);
  if (*piVar1 + 1 != 0) {
    *piVar1 = *piVar1 + 1;
  }
  lVar9 = FUN_18036cbb0(DAT_1803b1e90,piVar4,piVar1,0);
  FUN_180001d00(param_1,lVar9);
  plVar7 = (longlong *)FUN_1800015a0(DAT_1803ca7b8,DAT_1803b2570);
  if ((plVar7 == (longlong *)0x0) || (*plVar7 == 0)) {
    FUN_1800015a0(DAT_1803ab778,DAT_1803b2570);
  }
  cVar5 = FUN_18037f3a0();
  if (cVar5 == '\0') {
    plVar12 = *(longlong **)(param_1 + 0x60);
    uVar21 = 1;
    *(undefined8 *)(param_1 + 0x60) = 0;
  }
  else {
    FUN_1800b4e60();
    cVar5 = FUN_18037f3a0();
    if (cVar5 == '\0') {
      plVar12 = *(longlong **)(param_1 + 0x60);
      uVar21 = 1;
      *(undefined8 *)(param_1 + 0x60) = 0;
    }
    else {
      FUN_1800b4710();
      FUN_180372070();
      FUN_18037f0c0();
      FUN_1800b47a0();
      uVar3 = DAT_1803ca7b8;
      uVar2 = DAT_1803b2350;
      uVar17 = DAT_1803b2348;
      uVar8 = DAT_1803ab250;
      pcVar18 = _Py_NoneStruct_exref;
      uVar21 = 3;
      *(undefined4 *)(lVar9 + 0x28) = 3;
      plVar7 = (longlong *)FUN_1803717e0(param_1,uVar17,uVar3,pcVar18,uVar2,uVar8);
      if (plVar7 == (longlong *)0x0) {
        plVar12 = *(longlong **)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
      }
      else {
        if (((code *)plVar7[1] == PyModule_Type_exref) || (iVar19 = PyType_IsSubtype(), iVar19 != 0)
           ) {
          lVar10 = FUN_180371d80(param_1,plVar7,DAT_1803ca7b8,DAT_1803b2358,DAT_1803ab250);
        }
        else {
          lVar10 = FUN_180371890(param_1,plVar7,DAT_1803b2358);
        }
        if (lVar10 != 0) {
          FUN_1800b47a0();
          if (((code *)plVar7[1] == PyModule_Type_exref) ||
             (iVar19 = PyType_IsSubtype(), iVar19 != 0)) {
            lVar10 = FUN_180371d80(param_1,plVar7,DAT_1803ca7b8,DAT_1803b2360,DAT_1803ab250);
          }
          else {
            lVar10 = FUN_180371890(param_1,plVar7,DAT_1803b2360);
          }
          if (lVar10 != 0) {
            FUN_1800b47a0();
            if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
              (**(code **)(plVar7[1] + 0x30))(plVar7);
            }
            plVar7 = (longlong *)FUN_180372110();
            iVar19 = (int)*plVar7 + 1;
            if (iVar19 != 0) {
              *(int *)plVar7 = iVar19;
            }
            if (((code *)plVar7[1] == PyModule_Type_exref) ||
               (iVar19 = PyType_IsSubtype(), iVar19 != 0)) {
              lVar10 = FUN_180371d80(param_1,plVar7,DAT_1803ca7b8,DAT_1803b2368,DAT_1803ab250);
            }
            else {
              lVar10 = FUN_180371890(param_1,plVar7,DAT_1803b2368);
            }
            if (lVar10 != 0) {
              FUN_1800b47a0();
              if (((code *)plVar7[1] == PyModule_Type_exref) ||
                 (iVar19 = PyType_IsSubtype(), iVar19 != 0)) {
                lVar10 = FUN_180371d80(param_1,plVar7,DAT_1803ca7b8,DAT_1803b2370,DAT_1803ab250);
              }
              else {
                lVar10 = FUN_180371890(param_1,plVar7,DAT_1803b2370);
              }
              if (lVar10 != 0) {
                FUN_1800b47a0();
                if (((code *)plVar7[1] == PyModule_Type_exref) ||
                   (iVar19 = PyType_IsSubtype(), iVar19 != 0)) {
                  lVar10 = FUN_180371d80(param_1,plVar7,DAT_1803ca7b8,DAT_1803b2378,DAT_1803ab250);
                }
                else {
                  lVar10 = FUN_180371890(param_1,plVar7,DAT_1803b2378);
                }
                if (lVar10 != 0) {
                  FUN_1800b47a0();
                  if (((code *)plVar7[1] == PyModule_Type_exref) ||
                     (iVar19 = PyType_IsSubtype(), iVar19 != 0)) {
                    lVar10 = FUN_180371d80(param_1,plVar7,DAT_1803ca7b8,DAT_1803b2380,DAT_1803ab250)
                    ;
                  }
                  else {
                    lVar10 = FUN_180371890(param_1,plVar7,DAT_1803b2380);
                  }
                  if (lVar10 != 0) {
                    FUN_1800b47a0();
                    if (((code *)plVar7[1] == PyModule_Type_exref) ||
                       (iVar19 = PyType_IsSubtype(), iVar19 != 0)) {
                      lVar10 = FUN_180371d80(param_1,plVar7,DAT_1803ca7b8,DAT_1803b2388,
                                             DAT_1803ab250);
                    }
                    else {
                      lVar10 = FUN_180371890(param_1,plVar7,DAT_1803b2388);
                    }
                    if (lVar10 != 0) {
                      FUN_1800b47a0();
                      if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                        (**(code **)(plVar7[1] + 0x30))(plVar7);
                      }
                      uVar3 = DAT_1803ca7b8;
                      uVar2 = DAT_1803b23a0;
                      uVar17 = DAT_1803b2398;
                      uVar8 = DAT_1803b2390;
                      pcVar18 = _Py_NoneStruct_exref;
                      *(undefined4 *)(lVar9 + 0x28) = 6;
                      plVar7 = (longlong *)FUN_1803717e0(param_1,uVar8,uVar3,pcVar18,uVar17,uVar2);
                      if (plVar7 == (longlong *)0x0) {
                        plVar12 = *(longlong **)(param_1 + 0x60);
                        uVar21 = 6;
                        *(undefined8 *)(param_1 + 0x60) = 0;
                      }
                      else {
                        if (((code *)plVar7[1] == PyModule_Type_exref) ||
                           (iVar19 = PyType_IsSubtype(), iVar19 != 0)) {
                          lVar10 = FUN_180371d80(param_1,plVar7,DAT_1803ca7b8,DAT_1803b23a8,
                                                 DAT_1803ab250);
                        }
                        else {
                          lVar10 = FUN_180371890(param_1,plVar7,DAT_1803b23a8);
                        }
                        if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                          (**(code **)(plVar7[1] + 0x30))(plVar7);
                        }
                        if (lVar10 == 0) {
                          plVar12 = *(longlong **)(param_1 + 0x60);
                          uVar21 = 6;
                          *(undefined8 *)(param_1 + 0x60) = 0;
                        }
                        else {
                          FUN_1800b47a0();
                          uVar3 = DAT_1803ca7b8;
                          uVar2 = DAT_1803b23b8;
                          uVar17 = DAT_1803b23b0;
                          uVar8 = DAT_1803b23a0;
                          pcVar18 = _Py_NoneStruct_exref;
                          *(undefined4 *)(lVar9 + 0x28) = 7;
                          plVar7 = (longlong *)
                                   FUN_1803717e0(param_1,uVar17,uVar3,pcVar18,uVar2,uVar8);
                          if (plVar7 == (longlong *)0x0) {
                            plVar12 = *(longlong **)(param_1 + 0x60);
                            uVar21 = 7;
                            *(undefined8 *)(param_1 + 0x60) = 0;
                          }
                          else {
                            if (((code *)plVar7[1] == PyModule_Type_exref) ||
                               (iVar19 = PyType_IsSubtype(), iVar19 != 0)) {
                              lVar10 = FUN_180371d80(param_1,plVar7,DAT_1803ca7b8,DAT_1803b23c0,
                                                     DAT_1803ab250);
                            }
                            else {
                              lVar10 = FUN_180371890(param_1,plVar7,DAT_1803b23c0);
                            }
                            if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                              (**(code **)(plVar7[1] + 0x30))(plVar7);
                            }
                            if (lVar10 == 0) {
                              plVar12 = *(longlong **)(param_1 + 0x60);
                              uVar21 = 7;
                              *(undefined8 *)(param_1 + 0x60) = 0;
                            }
                            else {
                              FUN_1800b47a0();
                              uVar3 = DAT_1803ca7b8;
                              uVar2 = DAT_1803b23d0;
                              uVar17 = DAT_1803b23c8;
                              uVar8 = DAT_1803b23a0;
                              pcVar18 = _Py_NoneStruct_exref;
                              *(undefined4 *)(lVar9 + 0x28) = 8;
                              plVar7 = (longlong *)
                                       FUN_1803717e0(param_1,uVar17,uVar3,pcVar18,uVar2,uVar8);
                              if (plVar7 == (longlong *)0x0) {
                                plVar12 = *(longlong **)(param_1 + 0x60);
                                uVar21 = 8;
                                *(undefined8 *)(param_1 + 0x60) = 0;
                              }
                              else {
                                if (((code *)plVar7[1] == PyModule_Type_exref) ||
                                   (iVar19 = PyType_IsSubtype(), iVar19 != 0)) {
                                  lVar10 = FUN_180371d80(param_1,plVar7,DAT_1803ca7b8,DAT_1803b23d8,
                                                         DAT_1803ab250);
                                }
                                else {
                                  lVar10 = FUN_180371890(param_1,plVar7,DAT_1803b23d8);
                                }
                                if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                  (**(code **)(plVar7[1] + 0x30))(plVar7);
                                }
                                if (lVar10 == 0) {
                                  plVar12 = *(longlong **)(param_1 + 0x60);
                                  uVar21 = 8;
                                  *(undefined8 *)(param_1 + 0x60) = 0;
                                }
                                else {
                                  FUN_1800b47a0();
                                  FUN_1800b4710();
                                  FUN_1800b4710();
                                  FUN_1800b4710();
                                  plVar7 = (longlong *)FUN_180370ec0(param_1);
                                  plVar11 = (longlong *)FUN_180370ec0(param_1);
                                  lVar10 = FUN_1800b51e0();
                                  if (lVar10 == 0) {
                                    FUN_180373340(param_1,&local_res20,DAT_1803b2358);
                                    plVar12 = local_res20;
LAB_1800c2300:
                                    uVar21 = 0xf;
                                  }
                                  else {
                                    *(undefined4 *)(lVar9 + 0x28) = 0xf;
                                    plVar12 = (longlong *)
                                              FUN_18037ccb0(param_1,lVar10,DAT_1803b2400 + 0x18,
                                                            DAT_1803b2408);
                                    if (plVar12 == (longlong *)0x0) {
LAB_1800c22f8:
                                      plVar12 = *(longlong **)(param_1 + 0x60);
                                      *(undefined8 *)(param_1 + 0x60) = 0;
                                      goto LAB_1800c2300;
                                    }
                                    iVar19 = (int)*plVar11 + 1;
                                    if (iVar19 != 0) {
                                      *(int *)plVar11 = iVar19;
                                    }
                                    uVar8 = FUN_1800b4e10();
                                    PyDict_SetItem(plVar11,DAT_1803b2410,uVar8);
                                    PyDict_SetItem(plVar11,DAT_1803b2418,DAT_1803b22f8);
                                    plVar13 = (longlong *)FUN_180370ec0(param_1);
                                    PyDict_SetItem(plVar11,DAT_1803b2420,plVar13);
                                    if ((-1 < (int)*plVar13) &&
                                       (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                                      (**(code **)(plVar13[1] + 0x30))(plVar13);
                                    }
                                    lVar10 = FUN_18036cd80(param_1,DAT_1803b1f20,DAT_1803ca7c0);
                                    FUN_180001d00(param_1,lVar10);
                                    uVar8 = DAT_1803b2428;
                                    lVar14 = FUN_18036f670(param_1,plVar11,DAT_1803b2420);
                                    if (lVar14 == 0) {
                                      plVar12 = *(longlong **)(param_1 + 0x60);
                                      if (plVar12 != (longlong *)0x0) {
                                        plVar13 = (longlong *)plVar12[1];
                                        if ((*(uint *)((longlong *)plVar12[1] + 0x15) & 0x40000000)
                                            == 0) {
                                          plVar13 = plVar12;
                                        }
                                        if ((plVar13 != *(longlong **)PyExc_KeyError_exref) &&
                                           (((-1 < *(int *)(plVar13[1] + 0xa8) ||
                                             ((*(uint *)(plVar13 + 0x15) & 0x40000000) == 0)) ||
                                            (cVar5 = FUN_180382640(), cVar5 != '\x01')))) {
                                          uVar8 = 0x11;
                                          goto LAB_1800c2077;
                                        }
                                        *(undefined8 *)(param_1 + 0x60) = 0;
                                        if ((-1 < (int)*plVar12) &&
                                           (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                          (**(code **)(plVar12[1] + 0x30))(plVar12);
                                        }
                                      }
                                      FUN_180373340(param_1,&local_res20,DAT_1803b2420);
                                      uVar8 = 0x11;
                                      plVar12 = local_res20;
LAB_1800c207b:
                                      piVar1 = (int *)plVar12[5];
                                      if (piVar1 == (int *)0x0) {
                                        pcVar15 = (code *)FUN_18038e2a0(lVar10,uVar8);
                                        plVar13 = (longlong *)plVar12[5];
                                        pcVar18 = pcVar20;
                                        if (((pcVar15 != _Py_NoneStruct_exref) &&
                                            (pcVar18 = pcVar15, pcVar15 != (code *)0x0)) &&
                                           (*(int *)pcVar15 + 1 != 0)) {
                                          *(int *)pcVar15 = *(int *)pcVar15 + 1;
                                        }
                                        plVar12[5] = (longlong)pcVar18;
                                        if (((plVar13 != (longlong *)0x0) && (-1 < (int)*plVar13))
                                           && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                                          (**(code **)(plVar13[1] + 0x30))();
                                        }
                                        if (((pcVar15 != (code *)0x0) && (-1 < *(int *)pcVar15)) &&
                                           (*(longlong *)pcVar15 = *(longlong *)pcVar15 + -1,
                                           *(longlong *)pcVar15 == 0)) {
                                          (**(code **)(*(longlong *)(pcVar15 + 8) + 0x30))(pcVar15);
                                        }
                                      }
                                      else if (*(longlong *)(piVar1 + 6) != lVar10) {
                                        pcVar18 = (code *)FUN_18038e2a0(lVar10,uVar8);
                                        *(int **)(pcVar18 + 0x10) = piVar1;
                                        if (*piVar1 + 1 != 0) {
                                          *piVar1 = *piVar1 + 1;
                                        }
                                        plVar13 = (longlong *)plVar12[5];
                                        pcVar15 = pcVar20;
                                        if (((pcVar18 != _Py_NoneStruct_exref) &&
                                            (pcVar15 = pcVar18, pcVar18 != (code *)0x0)) &&
                                           (*(int *)pcVar18 + 1 != 0)) {
                                          *(int *)pcVar18 = *(int *)pcVar18 + 1;
                                        }
                                        plVar12[5] = (longlong)pcVar15;
                                        if (((plVar13 != (longlong *)0x0) && (-1 < (int)*plVar13))
                                           && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                                          (**(code **)(plVar13[1] + 0x30))();
                                        }
                                        if (((pcVar18 != (code *)0x0) && (-1 < *(int *)pcVar18)) &&
                                           (*(longlong *)pcVar18 = *(longlong *)pcVar18 + -1,
                                           *(longlong *)pcVar18 == 0)) {
                                          (**(code **)(*(longlong *)(pcVar18 + 8) + 0x30))(pcVar18);
                                        }
                                      }
                                      FUN_18036d0a0(lVar10,&DAT_180394940,0);
                                      FUN_180001d80(param_1);
                                    }
                                    else {
                                      cVar5 = FUN_18000d470(param_1,lVar14,DAT_1803b21e8,uVar8);
                                      uVar8 = DAT_1803b2430;
                                      if (cVar5 == '\0') {
                                        uVar8 = 0x11;
LAB_1800c2073:
                                        plVar12 = *(longlong **)(param_1 + 0x60);
LAB_1800c2077:
                                        *(undefined8 *)(param_1 + 0x60) = 0;
                                        goto LAB_1800c207b;
                                      }
                                      lVar14 = FUN_18036f670(param_1,plVar11,DAT_1803b2420);
                                      if (lVar14 == 0) {
                                        plVar12 = *(longlong **)(param_1 + 0x60);
                                        if (plVar12 != (longlong *)0x0) {
                                          plVar13 = (longlong *)plVar12[1];
                                          if ((*(uint *)((longlong *)plVar12[1] + 0x15) & 0x40000000
                                              ) == 0) {
                                            plVar13 = plVar12;
                                          }
                                          if ((plVar13 != *(longlong **)PyExc_KeyError_exref) &&
                                             (((-1 < *(int *)(plVar13[1] + 0xa8) ||
                                               ((*(uint *)(plVar13 + 0x15) & 0x40000000) == 0)) ||
                                              (cVar5 = FUN_180382640(), cVar5 != '\x01')))) {
                                            uVar8 = 0x12;
                                            goto LAB_1800c2077;
                                          }
                                          *(undefined8 *)(param_1 + 0x60) = 0;
                                          if ((-1 < (int)*plVar12) &&
                                             (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                            (**(code **)(plVar12[1] + 0x30))(plVar12);
                                          }
                                        }
                                        FUN_180373340(param_1,&local_res20,DAT_1803b2420);
                                        uVar8 = 0x12;
                                        plVar12 = local_res20;
                                        goto LAB_1800c207b;
                                      }
                                      cVar5 = FUN_18000d470(param_1,lVar14,DAT_1803b2108,uVar8);
                                      uVar8 = DAT_1803b2438;
                                      if (cVar5 == '\0') {
                                        uVar8 = 0x12;
                                        goto LAB_1800c2073;
                                      }
                                      lVar14 = FUN_18036f670(param_1,plVar11,DAT_1803b2420);
                                      if (lVar14 == 0) {
                                        plVar12 = *(longlong **)(param_1 + 0x60);
                                        if (plVar12 != (longlong *)0x0) {
                                          plVar13 = (longlong *)plVar12[1];
                                          if ((*(uint *)((longlong *)plVar12[1] + 0x15) & 0x40000000
                                              ) == 0) {
                                            plVar13 = plVar12;
                                          }
                                          if ((plVar13 != *(longlong **)PyExc_KeyError_exref) &&
                                             (((-1 < *(int *)(plVar13[1] + 0xa8) ||
                                               ((*(uint *)(plVar13 + 0x15) & 0x40000000) == 0)) ||
                                              (cVar5 = FUN_180382640(), cVar5 != '\x01')))) {
                                            uVar8 = 0x13;
                                            goto LAB_1800c2077;
                                          }
                                          *(undefined8 *)(param_1 + 0x60) = 0;
                                          if ((-1 < (int)*plVar12) &&
                                             (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                            (**(code **)(plVar12[1] + 0x30))(plVar12);
                                          }
                                        }
                                        FUN_180373340(param_1,&local_res20,DAT_1803b2420);
                                        uVar8 = 0x13;
                                        plVar12 = local_res20;
                                        goto LAB_1800c207b;
                                      }
                                      cVar5 = FUN_18000d470(param_1,lVar14,DAT_1803b20e8,uVar8);
                                      uVar8 = DAT_1803b2440;
                                      if (cVar5 == '\0') {
                                        uVar8 = 0x13;
                                        goto LAB_1800c2073;
                                      }
                                      lVar14 = FUN_18036f670(param_1,plVar11,DAT_1803b2420);
                                      if (lVar14 == 0) {
                                        plVar12 = *(longlong **)(param_1 + 0x60);
                                        if (plVar12 != (longlong *)0x0) {
                                          plVar13 = (longlong *)plVar12[1];
                                          if ((*(uint *)((longlong *)plVar12[1] + 0x15) & 0x40000000
                                              ) == 0) {
                                            plVar13 = plVar12;
                                          }
                                          if ((plVar13 != *(longlong **)PyExc_KeyError_exref) &&
                                             (((-1 < *(int *)(plVar13[1] + 0xa8) ||
                                               ((*(uint *)(plVar13 + 0x15) & 0x40000000) == 0)) ||
                                              (cVar5 = FUN_180382640(), cVar5 != '\x01')))) {
                                            uVar8 = 0x14;
                                            goto LAB_1800c2077;
                                          }
                                          *(undefined8 *)(param_1 + 0x60) = 0;
                                          if ((-1 < (int)*plVar12) &&
                                             (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                            (**(code **)(plVar12[1] + 0x30))(plVar12);
                                          }
                                        }
                                        FUN_180373340(param_1,&local_res20,DAT_1803b2420);
                                        uVar8 = 0x14;
                                        plVar12 = local_res20;
                                        goto LAB_1800c207b;
                                      }
                                      cVar5 = FUN_18000d470(param_1,lVar14,DAT_1803b2110,uVar8);
                                      uVar8 = DAT_1803b2448;
                                      if (cVar5 == '\0') {
                                        uVar8 = 0x14;
                                        goto LAB_1800c2073;
                                      }
                                      lVar14 = FUN_18036f670(param_1,plVar11,DAT_1803b2420);
                                      if (lVar14 == 0) {
                                        plVar12 = *(longlong **)(param_1 + 0x60);
                                        if (plVar12 != (longlong *)0x0) {
                                          plVar13 = (longlong *)plVar12[1];
                                          if ((*(uint *)((longlong *)plVar12[1] + 0x15) & 0x40000000
                                              ) == 0) {
                                            plVar13 = plVar12;
                                          }
                                          if ((plVar13 != *(longlong **)PyExc_KeyError_exref) &&
                                             (((-1 < *(int *)(plVar13[1] + 0xa8) ||
                                               ((*(uint *)(plVar13 + 0x15) & 0x40000000) == 0)) ||
                                              (cVar5 = FUN_180382640(), cVar5 != '\x01')))) {
                                            uVar8 = 0x15;
                                            goto LAB_1800c2077;
                                          }
                                          *(undefined8 *)(param_1 + 0x60) = 0;
                                          if ((-1 < (int)*plVar12) &&
                                             (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                            (**(code **)(plVar12[1] + 0x30))(plVar12);
                                          }
                                        }
                                        FUN_180373340(param_1,&local_res20,DAT_1803b2420);
                                        uVar8 = 0x15;
                                        plVar12 = local_res20;
                                        goto LAB_1800c207b;
                                      }
                                      cVar5 = FUN_18000d470(param_1,lVar14,DAT_1803b2118,uVar8);
                                      if (cVar5 == '\0') {
                                        uVar8 = 0x15;
                                        goto LAB_1800c2073;
                                      }
                                      FUN_180001d80(param_1);
                                      piVar1 = DAT_1803b22f8;
                                      pcVar18 = PyType_Type_exref;
                                      plVar13 = (longlong *)FUN_180374a40(param_1,3);
                                      if (*piVar1 + 1 != 0) {
                                        *piVar1 = *piVar1 + 1;
                                      }
                                      plVar13[3] = (longlong)piVar1;
                                      piVar1 = DAT_1803ab240;
                                      if (*DAT_1803ab240 + 1 != 0) {
                                        *DAT_1803ab240 = *DAT_1803ab240 + 1;
                                      }
                                      plVar13[4] = (longlong)piVar1;
                                      iVar19 = (int)*plVar11 + 1;
                                      if (iVar19 != 0) {
                                        *(int *)plVar11 = iVar19;
                                      }
                                      plVar13[5] = (longlong)plVar11;
                                      *(undefined4 *)(lVar9 + 0x28) = 0x10;
                                      plVar16 = (longlong *)
                                                FUN_180001790(param_1,pcVar18,plVar13,plVar7);
                                      if ((-1 < (int)*plVar13) &&
                                         (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                                        (**(code **)(plVar13[1] + 0x30))(plVar13);
                                      }
                                      if (plVar16 != (longlong *)0x0) {
                                        iVar19 = (int)*plVar16 + 1;
                                        if (iVar19 != 0) {
                                          *(int *)plVar16 = iVar19;
                                        }
                                        if ((-1 < (int)*plVar11) &&
                                           (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                          (**(code **)(plVar11[1] + 0x30))(plVar11);
                                        }
                                        if ((-1 < (int)*plVar16) &&
                                           (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                          (**(code **)(plVar16[1] + 0x30))(plVar16);
                                        }
                                        *(undefined4 *)(lVar9 + 0x28) = 0xf;
                                        lVar10 = FUN_180376770(param_1,plVar12,plVar16);
                                        if ((-1 < (int)*plVar12) &&
                                           (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                          (**(code **)(plVar12[1] + 0x30))(plVar12);
                                        }
                                        if ((-1 < (int)*plVar16) &&
                                           (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                          (**(code **)(plVar16[1] + 0x30))(plVar16);
                                        }
                                        if (lVar10 == 0) goto LAB_1800c22f8;
                                        FUN_1800b47a0();
                                        if ((-1 < (int)*plVar7) &&
                                           (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                          (**(code **)(plVar7[1] + 0x30))(plVar7);
                                        }
                                        if ((-1 < (int)*plVar11) &&
                                           (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                          (**(code **)(plVar11[1] + 0x30))(plVar11);
                                        }
                                        plVar7 = (longlong *)FUN_180370ec0(param_1);
                                        plVar11 = (longlong *)FUN_180370ec0(param_1);
                                        lVar10 = FUN_1800b51e0();
                                        if (lVar10 == 0) {
                                          FUN_180373340(param_1,&local_res20,DAT_1803b2358);
                                          plVar12 = local_res20;
LAB_1800c2eb6:
                                          uVar21 = 0x18;
                                        }
                                        else {
                                          *(undefined4 *)(lVar9 + 0x28) = 0x18;
                                          plVar12 = (longlong *)
                                                    FUN_18037ccb0(param_1,lVar10,
                                                                  DAT_1803b2400 + 0x18,DAT_1803b2408
                                                                 );
                                          if (plVar12 == (longlong *)0x0) {
LAB_1800c2eae:
                                            plVar12 = *(longlong **)(param_1 + 0x60);
                                            *(undefined8 *)(param_1 + 0x60) = 0;
                                            goto LAB_1800c2eb6;
                                          }
                                          iVar19 = (int)*plVar11 + 1;
                                          if (iVar19 != 0) {
                                            *(int *)plVar11 = iVar19;
                                          }
                                          uVar8 = FUN_1800b4e10();
                                          PyDict_SetItem(plVar11,DAT_1803b2410,uVar8);
                                          PyDict_SetItem(plVar11,DAT_1803b2418,DAT_1803b2100);
                                          plVar13 = (longlong *)FUN_180370ec0(param_1);
                                          PyDict_SetItem(plVar11,DAT_1803b2420,plVar13);
                                          if ((-1 < (int)*plVar13) &&
                                             (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                                            (**(code **)(plVar13[1] + 0x30))(plVar13);
                                          }
                                          lVar10 = FUN_18036cd80(param_1,DAT_1803b1ee0,DAT_1803ca7c0
                                                                );
                                          FUN_180001d00(param_1,lVar10);
                                          uVar8 = DAT_1803b2428;
                                          lVar14 = FUN_18036f670(param_1,plVar11,DAT_1803b2420);
                                          if (lVar14 == 0) {
                                            plVar12 = *(longlong **)(param_1 + 0x60);
                                            if (plVar12 != (longlong *)0x0) {
                                              plVar13 = (longlong *)plVar12[1];
                                              if ((*(uint *)((longlong *)plVar12[1] + 0x15) &
                                                  0x40000000) == 0) {
                                                plVar13 = plVar12;
                                              }
                                              if ((plVar13 != *(longlong **)PyExc_KeyError_exref) &&
                                                 (((-1 < *(int *)(plVar13[1] + 0xa8) ||
                                                   ((*(uint *)(plVar13 + 0x15) & 0x40000000) == 0))
                                                  || (cVar5 = FUN_180382640(), cVar5 != '\x01')))) {
                                                uVar8 = 0x1a;
LAB_1800c2cd9:
                                                *(undefined8 *)(param_1 + 0x60) = 0;
                                                local_res20 = plVar12;
                                                goto LAB_1800c2ce1;
                                              }
                                              *(undefined8 *)(param_1 + 0x60) = 0;
                                              if ((-1 < (int)*plVar12) &&
                                                 (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                                (**(code **)(plVar12[1] + 0x30))(plVar12);
                                              }
                                            }
                                            FUN_180373340(param_1,&local_res20,DAT_1803b2420);
                                            uVar8 = 0x1a;
LAB_1800c2ce1:
                                            plVar12 = local_res20;
                                            piVar1 = (int *)local_res20[5];
                                            if (piVar1 == (int *)0x0) {
                                              uVar8 = FUN_18038e2a0(lVar10,uVar8);
                                              FUN_180001300(&local_res20,uVar8);
                                            }
                                            else if (*(longlong *)(piVar1 + 6) != lVar10) {
                                              lVar14 = FUN_18038e2a0(lVar10,uVar8);
                                              *(int **)(lVar14 + 0x10) = piVar1;
                                              if (*piVar1 + 1 != 0) {
                                                *piVar1 = *piVar1 + 1;
                                              }
                                              FUN_180001300(&local_res20,lVar14);
                                            }
                                            FUN_18036d0a0(lVar10,&DAT_180394940,0);
                                            FUN_180001d80(param_1);
                                          }
                                          else {
                                            cVar5 = FUN_18000d470(param_1,lVar14,DAT_1803b2450,uVar8
                                                                 );
                                            uVar8 = DAT_1803b2430;
                                            if (cVar5 == '\0') {
                                              uVar8 = 0x1a;
                                              goto LAB_1800c2cd5;
                                            }
                                            lVar14 = FUN_18036f670(param_1,plVar11,DAT_1803b2420);
                                            if (lVar14 == 0) {
                                              plVar12 = *(longlong **)(param_1 + 0x60);
                                              if (plVar12 != (longlong *)0x0) {
                                                plVar13 = (longlong *)plVar12[1];
                                                if ((*(uint *)((longlong *)plVar12[1] + 0x15) &
                                                    0x40000000) == 0) {
                                                  plVar13 = plVar12;
                                                }
                                                if ((plVar13 != *(longlong **)PyExc_KeyError_exref)
                                                   && (((-1 < *(int *)(plVar13[1] + 0xa8) ||
                                                        ((*(uint *)(plVar13 + 0x15) & 0x40000000) ==
                                                         0)) || (cVar5 = FUN_180382640(),
                                                                cVar5 != '\x01')))) {
                                                  uVar8 = 0x1b;
                                                  goto LAB_1800c2cd9;
                                                }
                                                *(undefined8 *)(param_1 + 0x60) = 0;
                                                if ((-1 < (int)*plVar12) &&
                                                   (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                                  (**(code **)(plVar12[1] + 0x30))(plVar12);
                                                }
                                              }
                                              FUN_180373340(param_1,&local_res20,DAT_1803b2420);
                                              uVar8 = 0x1b;
                                              goto LAB_1800c2ce1;
                                            }
                                            cVar5 = FUN_18000d470(param_1,lVar14,DAT_1803b2458,uVar8
                                                                 );
                                            uVar8 = DAT_1803b2448;
                                            if (cVar5 == '\0') {
                                              uVar8 = 0x1b;
                                              goto LAB_1800c2cd5;
                                            }
                                            lVar14 = FUN_18036f670(param_1,plVar11,DAT_1803b2420);
                                            if (lVar14 == 0) {
                                              plVar12 = *(longlong **)(param_1 + 0x60);
                                              if (plVar12 != (longlong *)0x0) {
                                                plVar13 = (longlong *)plVar12[1];
                                                if ((*(uint *)((longlong *)plVar12[1] + 0x15) &
                                                    0x40000000) == 0) {
                                                  plVar13 = plVar12;
                                                }
                                                if ((plVar13 != *(longlong **)PyExc_KeyError_exref)
                                                   && (((-1 < *(int *)(plVar13[1] + 0xa8) ||
                                                        ((*(uint *)(plVar13 + 0x15) & 0x40000000) ==
                                                         0)) || (cVar5 = FUN_180382640(),
                                                                cVar5 != '\x01')))) {
                                                  uVar8 = 0x1c;
                                                  goto LAB_1800c2cd9;
                                                }
                                                *(undefined8 *)(param_1 + 0x60) = 0;
                                                if ((-1 < (int)*plVar12) &&
                                                   (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                                  (**(code **)(plVar12[1] + 0x30))(plVar12);
                                                }
                                              }
                                              FUN_180373340(param_1,&local_res20,DAT_1803b2420);
                                              uVar8 = 0x1c;
                                              goto LAB_1800c2ce1;
                                            }
                                            cVar5 = FUN_18000d470(param_1,lVar14,DAT_1803b2460,uVar8
                                                                 );
                                            uVar8 = DAT_1803b2438;
                                            if (cVar5 == '\0') {
                                              uVar8 = 0x1c;
                                              goto LAB_1800c2cd5;
                                            }
                                            lVar14 = FUN_18036f670(param_1,plVar11,DAT_1803b2420);
                                            if (lVar14 == 0) {
                                              plVar12 = *(longlong **)(param_1 + 0x60);
                                              if (plVar12 != (longlong *)0x0) {
                                                plVar13 = (longlong *)plVar12[1];
                                                if ((*(uint *)((longlong *)plVar12[1] + 0x15) &
                                                    0x40000000) == 0) {
                                                  plVar13 = plVar12;
                                                }
                                                if ((plVar13 != *(longlong **)PyExc_KeyError_exref)
                                                   && (((-1 < *(int *)(plVar13[1] + 0xa8) ||
                                                        ((*(uint *)(plVar13 + 0x15) & 0x40000000) ==
                                                         0)) || (cVar5 = FUN_180382640(),
                                                                cVar5 != '\x01')))) {
                                                  uVar8 = 0x1d;
                                                  goto LAB_1800c2cd9;
                                                }
                                                *(undefined8 *)(param_1 + 0x60) = 0;
                                                if ((-1 < (int)*plVar12) &&
                                                   (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                                  (**(code **)(plVar12[1] + 0x30))(plVar12);
                                                }
                                              }
                                              FUN_180373340(param_1,&local_res20,DAT_1803b2420);
                                              uVar8 = 0x1d;
                                              goto LAB_1800c2ce1;
                                            }
                                            cVar5 = FUN_18000d470(param_1,lVar14,DAT_1803b20e8,uVar8
                                                                 );
                                            uVar8 = DAT_1803b2440;
                                            if (cVar5 == '\0') {
                                              uVar8 = 0x1d;
                                              goto LAB_1800c2cd5;
                                            }
                                            lVar14 = FUN_18036f670(param_1,plVar11,DAT_1803b2420);
                                            if (lVar14 == 0) {
                                              plVar12 = *(longlong **)(param_1 + 0x60);
                                              if (plVar12 != (longlong *)0x0) {
                                                plVar13 = (longlong *)plVar12[1];
                                                if ((*(uint *)((longlong *)plVar12[1] + 0x15) &
                                                    0x40000000) == 0) {
                                                  plVar13 = plVar12;
                                                }
                                                if ((plVar13 != *(longlong **)PyExc_KeyError_exref)
                                                   && (((-1 < *(int *)(plVar13[1] + 0xa8) ||
                                                        ((*(uint *)(plVar13 + 0x15) & 0x40000000) ==
                                                         0)) || (cVar5 = FUN_180382640(),
                                                                cVar5 != '\x01')))) {
                                                  uVar8 = 0x1e;
                                                  goto LAB_1800c2cd9;
                                                }
                                                *(undefined8 *)(param_1 + 0x60) = 0;
                                                if ((-1 < (int)*plVar12) &&
                                                   (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                                  (**(code **)(plVar12[1] + 0x30))(plVar12);
                                                }
                                              }
                                              FUN_180373340(param_1,&local_res20,DAT_1803b2420);
                                              uVar8 = 0x1e;
                                              goto LAB_1800c2ce1;
                                            }
                                            cVar5 = FUN_18000d470(param_1,lVar14,DAT_1803b2110,uVar8
                                                                 );
                                            uVar8 = DAT_1803b2448;
                                            if (cVar5 == '\0') {
                                              uVar8 = 0x1e;
                                              goto LAB_1800c2cd5;
                                            }
                                            lVar14 = FUN_18036f670(param_1,plVar11,DAT_1803b2420);
                                            if (lVar14 == 0) {
                                              plVar12 = *(longlong **)(param_1 + 0x60);
                                              if (plVar12 != (longlong *)0x0) {
                                                plVar13 = (longlong *)plVar12[1];
                                                if ((*(uint *)((longlong *)plVar12[1] + 0x15) &
                                                    0x40000000) == 0) {
                                                  plVar13 = plVar12;
                                                }
                                                if ((plVar13 != *(longlong **)PyExc_KeyError_exref)
                                                   && (((-1 < *(int *)(plVar13[1] + 0xa8) ||
                                                        ((*(uint *)(plVar13 + 0x15) & 0x40000000) ==
                                                         0)) || (cVar5 = FUN_180382640(),
                                                                cVar5 != '\x01')))) {
                                                  uVar8 = 0x1f;
                                                  goto LAB_1800c2cd9;
                                                }
                                                *(undefined8 *)(param_1 + 0x60) = 0;
                                                if ((-1 < (int)*plVar12) &&
                                                   (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                                  (**(code **)(plVar12[1] + 0x30))(plVar12);
                                                }
                                              }
                                              FUN_180373340(param_1,&local_res20,DAT_1803b2420);
                                              uVar8 = 0x1f;
                                              goto LAB_1800c2ce1;
                                            }
                                            cVar5 = FUN_18000d470(param_1,lVar14,DAT_1803b2118,uVar8
                                                                 );
                                            if (cVar5 == '\0') {
                                              uVar8 = 0x1f;
                                              goto LAB_1800c2cd5;
                                            }
                                            PyDict_SetItem(plVar11,DAT_1803b2258,DAT_1803ab250);
                                            uVar8 = DAT_1803b2430;
                                            lVar14 = FUN_18036f670(param_1,plVar11,DAT_1803b2420);
                                            if (lVar14 == 0) {
                                              plVar12 = *(longlong **)(param_1 + 0x60);
                                              if (plVar12 != (longlong *)0x0) {
                                                plVar13 = (longlong *)plVar12[1];
                                                if ((*(uint *)((longlong *)plVar12[1] + 0x15) &
                                                    0x40000000) == 0) {
                                                  plVar13 = plVar12;
                                                }
                                                if ((plVar13 != *(longlong **)PyExc_KeyError_exref)
                                                   && (((-1 < *(int *)(plVar13[1] + 0xa8) ||
                                                        ((*(uint *)(plVar13 + 0x15) & 0x40000000) ==
                                                         0)) || (cVar5 = FUN_180382640(),
                                                                cVar5 != '\x01')))) {
                                                  uVar8 = 0x20;
                                                  goto LAB_1800c2cd9;
                                                }
                                                *(undefined8 *)(param_1 + 0x60) = 0;
                                                if ((-1 < (int)*plVar12) &&
                                                   (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                                  (**(code **)(plVar12[1] + 0x30))(plVar12);
                                                }
                                              }
                                              FUN_180373340(param_1,&local_res20,DAT_1803b2420);
                                              uVar8 = 0x20;
                                              goto LAB_1800c2ce1;
                                            }
                                            cVar5 = FUN_18000d470(param_1,lVar14,DAT_1803b2258,uVar8
                                                                 );
                                            if (cVar5 == '\0') {
                                              uVar8 = 0x20;
                                              goto LAB_1800c2cd5;
                                            }
                                            PyDict_SetItem(plVar11,DAT_1803b2468,
                                                           _Py_FalseStruct_exref);
                                            uVar8 = DAT_1803b2448;
                                            lVar14 = FUN_18036f670(param_1,plVar11,DAT_1803b2420);
                                            if (lVar14 == 0) {
                                              plVar12 = *(longlong **)(param_1 + 0x60);
                                              if (plVar12 != (longlong *)0x0) {
                                                plVar13 = (longlong *)plVar12[1];
                                                if ((*(uint *)((longlong *)plVar12[1] + 0x15) &
                                                    0x40000000) == 0) {
                                                  plVar13 = plVar12;
                                                }
                                                if ((plVar13 != *(longlong **)PyExc_KeyError_exref)
                                                   && (((-1 < *(int *)(plVar13[1] + 0xa8) ||
                                                        ((*(uint *)(plVar13 + 0x15) & 0x40000000) ==
                                                         0)) || (cVar5 = FUN_180382640(),
                                                                cVar5 != '\x01')))) {
                                                  uVar8 = 0x21;
                                                  goto LAB_1800c2cd9;
                                                }
                                                *(undefined8 *)(param_1 + 0x60) = 0;
                                                if ((-1 < (int)*plVar12) &&
                                                   (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                                  (**(code **)(plVar12[1] + 0x30))(plVar12);
                                                }
                                              }
                                              FUN_180373340(param_1,&local_res20,DAT_1803b2420);
                                              uVar8 = 0x21;
                                              goto LAB_1800c2ce1;
                                            }
                                            cVar5 = FUN_18000d470(param_1,lVar14,DAT_1803b2468,uVar8
                                                                 );
                                            if (cVar5 == '\0') {
                                              uVar8 = 0x21;
                                              goto LAB_1800c2cd5;
                                            }
                                            lVar14 = FUN_1800b5230(param_1);
                                            if (lVar14 == 0) {
                                              FUN_180373340(param_1,&local_res20,DAT_1803b2360);
                                              uVar8 = 0x22;
                                              goto LAB_1800c2ce1;
                                            }
                                            uVar8 = 0x22;
                                            *(undefined4 *)(lVar10 + 0x28) = 0x22;
                                            plVar13 = (longlong *)
                                                      FUN_18037ccb0(param_1,lVar14,
                                                                    DAT_1803b2470 + 0x18,
                                                                    DAT_1803b2478);
                                            if (plVar13 == (longlong *)0x0) {
LAB_1800c2cd5:
                                              plVar12 = *(longlong **)(param_1 + 0x60);
                                              goto LAB_1800c2cd9;
                                            }
                                            iVar19 = PyDict_SetItem(plVar11,DAT_1803b21f8,plVar13);
                                            if ((-1 < (int)*plVar13) &&
                                               (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                                              (**(code **)(plVar13[1] + 0x30))(plVar13);
                                            }
                                            uVar17 = DAT_1803b2480;
                                            if (iVar19 != 0) goto LAB_1800c2cd5;
                                            lVar14 = FUN_18036f670(param_1,plVar11,DAT_1803b2420);
                                            if (lVar14 == 0) {
                                              plVar12 = *(longlong **)(param_1 + 0x60);
                                              if (plVar12 != (longlong *)0x0) {
                                                plVar13 = (longlong *)plVar12[1];
                                                if ((*(uint *)((longlong *)plVar12[1] + 0x15) &
                                                    0x40000000) == 0) {
                                                  plVar13 = plVar12;
                                                }
                                                if ((plVar13 != *(longlong **)PyExc_KeyError_exref)
                                                   && (((-1 < *(int *)(plVar13[1] + 0xa8) ||
                                                        ((*(uint *)(plVar13 + 0x15) & 0x40000000) ==
                                                         0)) || (cVar5 = FUN_180382640(),
                                                                cVar5 != '\x01'))))
                                                goto LAB_1800c2cd9;
                                                *(undefined8 *)(param_1 + 0x60) = 0;
                                                if ((-1 < (int)*plVar12) &&
                                                   (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                                  (**(code **)(plVar12[1] + 0x30))(plVar12);
                                                }
                                              }
                                              FUN_180373340(param_1,&local_res20,DAT_1803b2420);
                                              goto LAB_1800c2ce1;
                                            }
                                            cVar5 = FUN_18000d470(param_1,lVar14,DAT_1803b21f8,
                                                                  uVar17);
                                            if (cVar5 == '\0') goto LAB_1800c2cd5;
                                            FUN_180001d80(param_1);
                                            piVar1 = DAT_1803b2100;
                                            pcVar18 = PyType_Type_exref;
                                            plVar13 = (longlong *)FUN_180374a40(param_1,3);
                                            if (*piVar1 + 1 != 0) {
                                              *piVar1 = *piVar1 + 1;
                                            }
                                            plVar13[3] = (longlong)piVar1;
                                            piVar1 = DAT_1803ab240;
                                            if (*DAT_1803ab240 + 1 != 0) {
                                              *DAT_1803ab240 = *DAT_1803ab240 + 1;
                                            }
                                            plVar13[4] = (longlong)piVar1;
                                            iVar19 = (int)*plVar11 + 1;
                                            if (iVar19 != 0) {
                                              *(int *)plVar11 = iVar19;
                                            }
                                            plVar13[5] = (longlong)plVar11;
                                            *(undefined4 *)(lVar9 + 0x28) = 0x19;
                                            plVar16 = (longlong *)
                                                      FUN_180001790(param_1,pcVar18,plVar13,plVar7);
                                            if ((-1 < (int)*plVar13) &&
                                               (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                                              (**(code **)(plVar13[1] + 0x30))(plVar13);
                                            }
                                            if (plVar16 != (longlong *)0x0) {
                                              iVar19 = (int)*plVar16 + 1;
                                              if (iVar19 != 0) {
                                                *(int *)plVar16 = iVar19;
                                              }
                                              if ((-1 < (int)*plVar11) &&
                                                 (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                (**(code **)(plVar11[1] + 0x30))(plVar11);
                                              }
                                              if ((-1 < (int)*plVar16) &&
                                                 (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                                (**(code **)(plVar16[1] + 0x30))(plVar16);
                                              }
                                              *(undefined4 *)(lVar9 + 0x28) = 0x18;
                                              lVar10 = FUN_180376770(param_1,plVar12,plVar16);
                                              if ((-1 < (int)*plVar12) &&
                                                 (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                                (**(code **)(plVar12[1] + 0x30))(plVar12);
                                              }
                                              if ((-1 < (int)*plVar16) &&
                                                 (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                                (**(code **)(plVar16[1] + 0x30))(plVar16);
                                              }
                                              if (lVar10 != 0) {
                                                FUN_1800b47a0();
                                                if ((-1 < (int)*plVar7) &&
                                                   (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                  (**(code **)(plVar7[1] + 0x30))(plVar7);
                                                }
                                                if ((-1 < (int)*plVar11) &&
                                                   (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                  (**(code **)(plVar11[1] + 0x30))(plVar11);
                                                }
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803b2488);
                                                FUN_1803609a0(FUN_1800b5770,DAT_1803b22d8,0,
                                                              DAT_1803b1e70,0,0,uVar8,DAT_1803ca7c0,
                                                              0,0,0);
                                                FUN_1800b47a0();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803b2490);
                                                uVar17 = FUN_18036fb80(param_1,DAT_1803b2498);
                                                FUN_1803609a0(FUN_1800b5cc0,DAT_1803b22e8,0,
                                                              DAT_1803b1f40,0,uVar8,uVar17,
                                                              DAT_1803ca7c0,0,0,0);
                                                FUN_1800b47a0();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803b24a0);
                                                FUN_1803609a0(FUN_1800b61f0,DAT_1803b22e0,0,
                                                              DAT_1803b1f28,0,0,uVar8,DAT_1803ca7c0,
                                                              0,0,0);
                                                FUN_1800b47a0();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803b24a8);
                                                FUN_1803609a0(FUN_1800b67e0,DAT_1803b2128,0,
                                                              DAT_1803b2578,0,0,uVar8,DAT_1803ca7c0,
                                                              0,0,0);
                                                FUN_1800b47a0();
                                                piVar1 = DAT_1803b24b0;
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803b24b8);
                                                if (*piVar1 + 1 != 0) {
                                                  *piVar1 = *piVar1 + 1;
                                                }
                                                FUN_1803609a0(FUN_1800b6c70,DAT_1803b21b8,0,
                                                              DAT_1803b1eb0,piVar1,0,uVar8,
                                                              DAT_1803ca7c0,0,0,0);
                                                FUN_1800b47a0();
                                                FUN_18036fb80(param_1,DAT_1803b24c0);
                                                FUN_1800c1080();
                                                FUN_1800b47a0();
                                                uVar8 = DAT_1803b24c8;
                                                plVar7 = (longlong *)
                                                         FUN_1800015a0(DAT_1803ca7b8,DAT_1803b23f8);
                                                if (((plVar7 == (longlong *)0x0) ||
                                                    (lVar10 = *plVar7, lVar10 == 0)) &&
                                                   ((plVar7 = (longlong *)
                                                              FUN_1800015a0(DAT_1803ab778,
                                                                            DAT_1803b23f8),
                                                    plVar7 == (longlong *)0x0 ||
                                                    (lVar10 = *plVar7, lVar10 == 0)))) {
                                                  FUN_180373340(param_1,&local_res20,DAT_1803b23f8);
                                                  uVar21 = 0x84;
                                                  plVar12 = local_res20;
                                                }
                                                else {
                                                  plVar7 = (longlong *)_PyDict_NewPresized(9);
                                                  PyDict_SetItem(plVar7,uVar8,lVar10);
                                                  uVar8 = DAT_1803b2200;
                                                  plVar12 = (longlong *)
                                                            FUN_1800015a0(DAT_1803ca7b8,
                                                                          DAT_1803b23e8);
                                                  if (((plVar12 != (longlong *)0x0) &&
                                                      (*plVar12 != 0)) ||
                                                     ((plVar12 = (longlong *)
                                                                 FUN_1800015a0(DAT_1803ab778,
                                                                               DAT_1803b23e8),
                                                      plVar12 != (longlong *)0x0 && (*plVar12 != 0))
                                                     )) {
                                                    PyDict_SetItem(plVar7,uVar8);
                                                    PyDict_SetItem(plVar7,DAT_1803b2280,
                                                                   _Py_NoneStruct_exref);
                                                    PyDict_SetItem(plVar7,DAT_1803b2150,
                                                                   _Py_NoneStruct_exref);
                                                    PyDict_SetItem(plVar7,DAT_1803b2290,
                                                                   _Py_NoneStruct_exref);
                                                    PyDict_SetItem(plVar7,DAT_1803b2298,
                                                                   _Py_NoneStruct_exref);
                                                    PyDict_SetItem(plVar7,DAT_1803b22a0,
                                                                   _Py_NoneStruct_exref);
                                                    PyDict_SetItem(plVar7,DAT_1803b2158,
                                                                   DAT_1803b24d0);
                                                    PyDict_SetItem(plVar7,DAT_1803b2170,
                                                                   DAT_1803b24d8);
                                                    uVar8 = FUN_18036fb80(param_1,DAT_1803b24e0);
                                                    FUN_1803609a0(FUN_1800b9330,DAT_1803b24e8,0,
                                                                  DAT_1803b1e98,0,plVar7,uVar8,
                                                                  DAT_1803ca7c0,0,0,0);
                                                    FUN_1800b47a0();
                                                    FUN_180001d80(param_1);
                                                    piVar1 = DAT_1803ca7c0;
                                                    if (*DAT_1803ca7c0 + 1 == 0) {
                                                      return DAT_1803ca7c0;
                                                    }
                                                    *DAT_1803ca7c0 = *DAT_1803ca7c0 + 1;
                                                    return piVar1;
                                                  }
                                                  FUN_180373340(param_1,&local_res20,DAT_1803b23e8);
                                                  uVar21 = 0x85;
                                                  plVar12 = local_res20;
                                                  if ((-1 < (int)*plVar7) &&
                                                     (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                    (**(code **)(plVar7[1] + 0x30))(plVar7);
                                                    plVar12 = local_res20;
                                                  }
                                                }
                                                goto LAB_1800c33be;
                                              }
                                              goto LAB_1800c2eae;
                                            }
                                            plVar12 = *(longlong **)(param_1 + 0x60);
                                            *(undefined8 *)(param_1 + 0x60) = 0;
                                          }
                                          if ((-1 < (int)*plVar11) &&
                                             (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                            (**(code **)(plVar11[1] + 0x30))(plVar11);
                                          }
                                          uVar21 = 0x19;
                                        }
                                        if ((-1 < (int)*plVar7) &&
                                           (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                          (**(code **)(plVar7[1] + 0x30))(plVar7);
                                        }
                                        if ((-1 < (int)*plVar11) &&
                                           (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                          (**(code **)(plVar11[1] + 0x30))(plVar11);
                                        }
                                        goto LAB_1800c33be;
                                      }
                                      plVar12 = *(longlong **)(param_1 + 0x60);
                                      *(undefined8 *)(param_1 + 0x60) = 0;
                                    }
                                    if ((-1 < (int)*plVar11) &&
                                       (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                      (**(code **)(plVar11[1] + 0x30))(plVar11);
                                    }
                                    uVar21 = 0x10;
                                  }
                                  if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0))
                                  {
                                    (**(code **)(plVar7[1] + 0x30))(plVar7);
                                  }
                                  if ((-1 < (int)*plVar11) &&
                                     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                    (**(code **)(plVar11[1] + 0x30))(plVar11);
                                  }
                                }
                              }
                            }
                          }
                        }
                      }
                      goto LAB_1800c33be;
                    }
                  }
                }
              }
            }
            uVar21 = 4;
          }
        }
        plVar12 = *(longlong **)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
        if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
          (**(code **)(plVar7[1] + 0x30))(plVar7);
        }
      }
    }
  }
LAB_1800c33be:
  piVar1 = (int *)plVar12[5];
  if (piVar1 == (int *)0x0) {
    pcVar18 = (code *)FUN_18038e2a0(lVar9,uVar21);
    bVar22 = pcVar18 == _Py_NoneStruct_exref;
  }
  else {
    if (*(longlong *)(piVar1 + 6) == lVar9) goto LAB_1800c345c;
    pcVar18 = (code *)FUN_18038e2a0(lVar9,uVar21);
    *(int **)(pcVar18 + 0x10) = piVar1;
    if (*piVar1 + 1 != 0) {
      *piVar1 = *piVar1 + 1;
    }
    bVar22 = pcVar18 == _Py_NoneStruct_exref;
  }
  plVar7 = (longlong *)plVar12[5];
  if (((!bVar22) && (pcVar20 = pcVar18, pcVar18 != (code *)0x0)) && (*(int *)pcVar18 + 1 != 0)) {
    *(int *)pcVar18 = *(int *)pcVar18 + 1;
  }
  plVar12[5] = (longlong)pcVar20;
  if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)
     ) {
    (**(code **)(plVar7[1] + 0x30))();
  }
  if (((pcVar18 != (code *)0x0) && (-1 < *(int *)pcVar18)) &&
     (*(longlong *)pcVar18 = *(longlong *)pcVar18 + -1, *(longlong *)pcVar18 == 0)) {
    (**(code **)(*(longlong *)(pcVar18 + 8) + 0x30))(pcVar18);
  }
LAB_1800c345c:
  FUN_180001d80(param_1);
  plVar7 = *(longlong **)(param_1 + 0x60);
  *(longlong **)(param_1 + 0x60) = plVar12;
  if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)
     ) {
    (**(code **)(plVar7[1] + 0x30))();
  }
  return (int *)0x0;
}

