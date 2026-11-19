/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x18004dc60 */

longlong * FUN_18004dc60(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  undefined8 *puVar2;
  int *piVar3;
  undefined8 uVar4;
  char cVar5;
  int iVar6;
  longlong *plVar7;
  longlong lVar8;
  longlong *plVar9;
  longlong *plVar10;
  longlong *plVar11;
  longlong *plVar12;
  longlong *plVar13;
  longlong *plVar14;
  code *pcVar15;
  code *pcVar16;
  longlong *plVar17;
  undefined8 uVar18;
  bool bVar19;
  longlong *local_68;
  longlong *local_60;
  longlong *local_58;
  
  plVar1 = (longlong *)*param_3;
  plVar17 = (longlong *)0x0;
  lVar8 = *(longlong *)(param_1 + 0x10);
  local_68 = (longlong *)0x0;
  local_60 = (longlong *)0x0;
  local_58 = (longlong *)0x0;
  iVar6 = *(int *)(lVar8 + 0x42108);
  if (iVar6 == 0) {
    plVar7 = (longlong *)FUN_180001220(PyList_Type_exref);
  }
  else {
    *(int *)(lVar8 + 0x42108) = iVar6 + -1;
    plVar7 = *(longlong **)(lVar8 + 0x41e80 + (longlong)iVar6 * 8);
    *plVar7 = 1;
  }
  plVar10 = plVar7 + -2;
  plVar7[3] = 0;
  plVar7[2] = 0;
  plVar7[4] = 0;
  lVar8 = _PyThreadState_GetCurrent();
  plVar9 = DAT_1803aec48;
  lVar8 = *(longlong *)(*(longlong *)(lVar8 + 0x10) + 0xd0);
  puVar2 = *(undefined8 **)(lVar8 + 8);
  *puVar2 = plVar10;
  plVar7[-1] = plVar7[-1] & 3;
  plVar7[-1] = plVar7[-1] | (ulonglong)puVar2;
  *plVar10 = lVar8;
  *(longlong **)(lVar8 + 8) = plVar10;
  if (plVar9 == (longlong *)0x0) {
LAB_18004dd52:
    plVar9 = (longlong *)FUN_18036cd10(param_1,DAT_1803ae760,DAT_1803ca880,0x28);
    DAT_1803aec48 = plVar9;
  }
  else if ((1 < *plVar9) || (plVar9[2] != 0)) {
    if ((plVar9 != (longlong *)0x0) &&
       ((-1 < (int)*plVar9 && (*plVar9 = *plVar9 + -1, *plVar9 == 0)))) {
      (**(code **)(plVar9[1] + 0x30))(plVar9);
    }
    goto LAB_18004dd52;
  }
  FUN_180001d00(param_1,plVar9);
  plVar10 = (longlong *)FUN_18037f0c0();
  if (plVar10 == (longlong *)0x0) {
    plVar11 = *(longlong **)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
    uVar18 = 0x91;
  }
  else {
    uVar18 = 0x91;
    *(undefined4 *)(plVar9 + 5) = 0x91;
    plVar11 = (longlong *)FUN_180376210(param_1,plVar10);
    if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
      (**(code **)(plVar10[1] + 0x30))(plVar10);
    }
    if (plVar11 == (longlong *)0x0) {
      plVar11 = *(longlong **)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      plVar17 = (longlong *)0x0;
    }
    else {
      plVar10 = (longlong *)FUN_18000d280();
      if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
        (**(code **)(plVar11[1] + 0x30))(plVar11);
      }
      if (plVar10 == (longlong *)0x0) {
        plVar11 = *(longlong **)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
        plVar17 = (longlong *)0x0;
      }
      else {
        do {
          do {
            pcVar15 = *(code **)(plVar10[1] + 0xe0);
            if (pcVar15 == (code *)0x0) {
              PyErr_Format(*(undefined8 *)PyExc_TypeError_exref,"\'%s\' object is not an iterator",
                           *(undefined8 *)(plVar10[1] + 0x18));
LAB_18004e4ca:
              plVar11 = *(longlong **)(param_1 + 0x60);
              local_68 = local_58;
              if (plVar11 != (longlong *)0x0) {
                plVar12 = (longlong *)plVar11[1];
                if ((*(uint *)((longlong *)plVar11[1] + 0x15) & 0x40000000) == 0) {
                  plVar12 = plVar11;
                }
                if ((plVar12 != *(longlong **)PyExc_StopIteration_exref) &&
                   (((-1 < *(int *)(plVar12[1] + 0xa8) ||
                     ((*(uint *)(plVar12 + 0x15) & 0x40000000) == 0)) ||
                    (cVar5 = FUN_180382640(), cVar5 != '\x01')))) {
                  *(undefined8 *)(param_1 + 0x60) = 0;
                  uVar18 = 0x91;
                  if (local_58 == (longlong *)0x0) goto LAB_18004e568;
                  goto LAB_18004e549;
                }
                *(undefined8 *)(param_1 + 0x60) = 0;
                if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                  (**(code **)(plVar11[1] + 0x30))(plVar11);
                }
              }
              if (((local_58 != (longlong *)0x0) && (-1 < (int)*local_58)) &&
                 (*local_58 = *local_58 + -1, *local_58 == 0)) {
                (**(code **)(local_58[1] + 0x30))(local_58);
              }
              if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                (**(code **)(plVar10[1] + 0x30))();
              }
              iVar6 = FUN_180001510(plVar7);
              if (iVar6 == -1) {
                plVar11 = *(longlong **)(param_1 + 0x60);
                uVar18 = 0xa0;
                *(undefined8 *)(param_1 + 0x60) = 0;
                goto LAB_18004e592;
              }
              bVar19 = iVar6 == 0;
              if (iVar6 == 0) {
                plVar10 = (longlong *)FUN_18037f0c0();
                uVar18 = 0xa0;
                if (plVar10 == (longlong *)0x0) {
                  plVar11 = *(longlong **)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                  goto LAB_18004e592;
                }
                *(undefined4 *)(plVar9 + 5) = 0xa0;
                plVar12 = (longlong *)FUN_180376210(param_1,plVar10);
                if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                  (**(code **)(plVar10[1] + 0x30))(plVar10);
                }
                if (plVar12 == (longlong *)0x0) {
                  plVar11 = *(longlong **)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                  goto LAB_18004e592;
                }
                iVar6 = FUN_180001510(plVar12);
                if (iVar6 == -1) {
                  plVar11 = *(longlong **)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                  if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                    (**(code **)(plVar12[1] + 0x30))(plVar12);
                  }
                  goto LAB_18004e592;
                }
                bVar19 = iVar6 != 0;
                if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                  (**(code **)(plVar12[1] + 0x30))(plVar12);
                }
              }
              uVar18 = DAT_1803ae9f0;
              plVar10 = DAT_1803ae9b8;
              if (!bVar19) {
LAB_18004ea44:
                iVar6 = (int)*plVar7 + 1;
                if (iVar6 != 0) {
                  *(int *)plVar7 = iVar6;
                }
                FUN_180001d80(param_1);
                if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                  (**(code **)(plVar7[1] + 0x30))(plVar7);
                }
                if (((local_58 != (longlong *)0x0) && (-1 < (int)*local_58)) &&
                   (*local_58 = *local_58 + -1, *local_58 == 0)) {
                  (**(code **)(local_58[1] + 0x30))();
                }
                if (((local_60 != (longlong *)0x0) && (-1 < (int)*local_60)) &&
                   (*local_60 = *local_60 + -1, *local_60 == 0)) {
                  (**(code **)(local_60[1] + 0x30))(local_60);
                }
                if (((plVar17 != (longlong *)0x0) && (-1 < (int)*plVar17)) &&
                   (*plVar17 = *plVar17 + -1, *plVar17 == 0)) {
                  (**(code **)(plVar17[1] + 0x30))(plVar17);
                }
                if ((int)*plVar1 < 0) {
                  return plVar7;
                }
                *plVar1 = *plVar1 + -1;
                if (*plVar1 != 0) {
                  return plVar7;
                }
                (**(code **)(plVar1[1] + 0x30))();
                return plVar7;
              }
              plVar12 = (longlong *)_PyDict_NewPresized(4);
              PyDict_SetItem(plVar12,uVar18,plVar10);
              PyDict_SetItem(plVar12,DAT_1803ae9f8,DAT_1803aea00);
              uVar4 = DAT_1803aea08;
              plVar10 = (longlong *)FUN_18037f0c0();
              uVar18 = 0xa4;
              if (plVar10 != (longlong *)0x0) {
                *(undefined4 *)(plVar9 + 5) = 0xa4;
                plVar11 = (longlong *)FUN_180376210(param_1,plVar10);
                if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                  (**(code **)(plVar10[1] + 0x30))(plVar10);
                }
                if (plVar11 != (longlong *)0x0) {
                  PyDict_SetItem(plVar12,uVar4,plVar11);
                  if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                    (**(code **)(plVar11[1] + 0x30))(plVar11);
                  }
                  PyDict_SetItem(plVar12,DAT_1803aea20,DAT_1803aea38);
                  FUN_180372560(plVar7,plVar12);
                  goto LAB_18004ea44;
                }
              }
              plVar11 = *(longlong **)(param_1 + 0x60);
              *(undefined8 *)(param_1 + 0x60) = 0;
              if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                (**(code **)(plVar12[1] + 0x30))(plVar12);
              }
              goto LAB_18004e592;
            }
            plVar11 = (longlong *)(*pcVar15)();
            if (plVar11 == (longlong *)0x0) goto LAB_18004e4ca;
            if (((local_58 != (longlong *)0x0) && (-1 < (int)*local_58)) &&
               (*local_58 = *local_58 + -1, *local_58 == 0)) {
              (**(code **)(local_58[1] + 0x30))(local_58);
            }
            iVar6 = (int)*plVar11 + 1;
            if (iVar6 != 0) {
              *(int *)plVar11 = iVar6;
            }
            if (((local_58 != (longlong *)0x0) && (-1 < (int)*local_58)) &&
               (*local_58 = *local_58 + -1, *local_58 == 0)) {
              (**(code **)(local_58[1] + 0x30))();
            }
            plVar12 = (longlong *)FUN_18037f0c0();
            local_58 = plVar11;
            if (plVar12 == (longlong *)0x0) {
LAB_18004e493:
              plVar11 = *(longlong **)(param_1 + 0x60);
              uVar18 = 0x92;
              *(undefined8 *)(param_1 + 0x60) = 0;
              goto LAB_18004e549;
            }
            *(undefined4 *)(plVar9 + 5) = 0x92;
            plVar11 = (longlong *)FUN_180376210(param_1,plVar12);
            if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
              (**(code **)(plVar12[1] + 0x30))(plVar12);
            }
            if (plVar11 == (longlong *)0x0) goto LAB_18004e493;
            if (((local_60 != (longlong *)0x0) && (-1 < (int)*local_60)) &&
               (*local_60 = *local_60 + -1, *local_60 == 0)) {
              (**(code **)(local_60[1] + 0x30))(local_60);
            }
            plVar12 = (longlong *)FUN_18037f0c0();
            local_60 = plVar11;
            if (plVar12 == (longlong *)0x0) {
LAB_18004e47c:
              plVar11 = *(longlong **)(param_1 + 0x60);
              uVar18 = 0x93;
              *(undefined8 *)(param_1 + 0x60) = 0;
              goto LAB_18004e549;
            }
            *(undefined4 *)(plVar9 + 5) = 0x93;
            plVar13 = (longlong *)FUN_1803770d0(param_1,plVar12,DAT_1803ae9b0);
            if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
              (**(code **)(plVar12[1] + 0x30))(plVar12);
            }
            if (plVar13 == (longlong *)0x0) goto LAB_18004e47c;
            iVar6 = FUN_180001510(plVar13);
            if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
              (**(code **)(plVar13[1] + 0x30))(plVar13);
            }
            plVar12 = DAT_1803ae9b8;
            if (iVar6 == -1) goto LAB_18004e47c;
          } while (iVar6 == 0);
          iVar6 = (int)*DAT_1803ae9b8 + 1;
          if (iVar6 != 0) {
            *(int *)DAT_1803ae9b8 = iVar6;
          }
          if (((plVar17 != (longlong *)0x0) && (-1 < (int)*plVar17)) &&
             (*plVar17 = *plVar17 + -1, *plVar17 == 0)) {
            (**(code **)(plVar17[1] + 0x30))();
          }
          uVar18 = DAT_1803ae9c0;
          plVar13 = (longlong *)FUN_18037f0c0();
          plVar17 = plVar12;
          if (plVar13 == (longlong *)0x0) {
LAB_18004e465:
            plVar11 = *(longlong **)(param_1 + 0x60);
            uVar18 = 0x96;
            *(undefined8 *)(param_1 + 0x60) = 0;
            goto LAB_18004e549;
          }
          *(undefined4 *)(plVar9 + 5) = 0x96;
          plVar14 = (longlong *)FUN_180376210(param_1,plVar13);
          if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
            (**(code **)(plVar13[1] + 0x30))(plVar13);
          }
          if (plVar14 == (longlong *)0x0) goto LAB_18004e465;
          iVar6 = PySequence_Contains(plVar14,uVar18);
          if ((-1 < (int)*plVar14) && (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
            (**(code **)(plVar14[1] + 0x30))(plVar14);
          }
          if (iVar6 == -1) goto LAB_18004e465;
          bVar19 = iVar6 == 1;
          if (!bVar19) {
            iVar6 = PySequence_Contains(plVar11,DAT_1803ae9c8);
            if (iVar6 == -1) goto LAB_18004e465;
            bVar19 = iVar6 == 1;
          }
          uVar18 = DAT_1803ae9d8;
          plVar13 = DAT_1803ae9d0;
          if (!bVar19) {
            plVar13 = (longlong *)FUN_18037f0c0();
            if (plVar13 != (longlong *)0x0) {
              *(undefined4 *)(plVar9 + 5) = 0x98;
              plVar14 = (longlong *)FUN_180376210(param_1,plVar13);
              if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                (**(code **)(plVar13[1] + 0x30))(plVar13);
              }
              if (plVar14 != (longlong *)0x0) {
                iVar6 = PySequence_Contains(plVar14,uVar18);
                if ((-1 < (int)*plVar14) && (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                  (**(code **)(plVar14[1] + 0x30))(plVar14);
                }
                if (iVar6 != -1) {
                  bVar19 = iVar6 == 1;
                  if (!bVar19) {
                    iVar6 = PySequence_Contains(plVar11,DAT_1803ae9e0);
                    if (iVar6 == -1) goto LAB_18004e44e;
                    bVar19 = iVar6 == 1;
                  }
                  plVar13 = DAT_1803ae9b8;
                  if (bVar19) {
                    iVar6 = (int)*DAT_1803ae9b8 + 1;
                    if (iVar6 != 0) {
                      *(int *)DAT_1803ae9b8 = iVar6;
                    }
                    plVar17 = plVar13;
                    if (-1 < (int)*plVar12) {
                      *plVar12 = *plVar12 + -1;
                      lVar8 = *plVar12;
                      goto joined_r0x00018004e1d3;
                    }
                  }
                  goto LAB_18004e1df;
                }
              }
            }
LAB_18004e44e:
            plVar11 = *(longlong **)(param_1 + 0x60);
            uVar18 = 0x98;
            *(undefined8 *)(param_1 + 0x60) = 0;
            goto LAB_18004e549;
          }
          iVar6 = (int)*DAT_1803ae9d0 + 1;
          if (iVar6 != 0) {
            *(int *)DAT_1803ae9d0 = iVar6;
          }
          plVar17 = plVar13;
          if (-1 < (int)*plVar12) {
            *plVar12 = *plVar12 + -1;
            lVar8 = *plVar12;
joined_r0x00018004e1d3:
            plVar17 = plVar13;
            if (lVar8 == 0) {
              (**(code **)(plVar12[1] + 0x30))(plVar12);
            }
          }
LAB_18004e1df:
          uVar18 = DAT_1803ae9f0;
          plVar12 = (longlong *)_PyDict_NewPresized(4);
          PyDict_SetItem(plVar12,uVar18,plVar17);
          PyDict_SetItem(plVar12,DAT_1803ae9f8,DAT_1803aea00);
          uVar18 = DAT_1803aea08;
          plVar11 = (longlong *)FUN_18037f0c0();
          if (plVar11 == (longlong *)0x0) {
LAB_18004e416:
            plVar11 = *(longlong **)(param_1 + 0x60);
            uVar18 = 0x9d;
            *(undefined8 *)(param_1 + 0x60) = 0;
            if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
              (**(code **)(plVar12[1] + 0x30))(plVar12);
            }
            goto LAB_18004e549;
          }
          *(undefined4 *)(plVar9 + 5) = 0x9d;
          plVar13 = (longlong *)FUN_1803770d0(param_1,plVar11,DAT_1803aea18);
          if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
            (**(code **)(plVar11[1] + 0x30))(plVar11);
          }
          if (plVar13 == (longlong *)0x0) goto LAB_18004e416;
          plVar11 = (longlong *)FUN_18037f0c0();
          if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
            (**(code **)(plVar13[1] + 0x30))(plVar13);
          }
          if (plVar11 == (longlong *)0x0) goto LAB_18004e416;
          *(undefined4 *)(plVar9 + 5) = 0x9d;
          plVar13 = (longlong *)FUN_180376210(param_1,plVar11);
          if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
            (**(code **)(plVar11[1] + 0x30))(plVar11);
          }
          if (plVar13 == (longlong *)0x0) goto LAB_18004e416;
          PyDict_SetItem(plVar12,uVar18,plVar13);
          if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
            (**(code **)(plVar13[1] + 0x30))(plVar13);
          }
          cVar5 = FUN_180383620(plVar17,DAT_1803ae9d0);
          uVar18 = DAT_1803aea30;
          if (cVar5 != '\0') {
            uVar18 = DAT_1803aea28;
          }
          PyDict_SetItem(plVar12,DAT_1803aea20,uVar18);
          lVar8 = plVar7[2];
          cVar5 = FUN_1803722b0(plVar7,lVar8 + 1);
          if (cVar5 != '\0') {
            *(longlong **)(plVar7[3] + lVar8 * 8) = plVar12;
          }
          lVar8 = *(longlong *)(param_1 + 0x10);
          if ((((*(int *)(*(longlong *)(lVar8 + 0x60) + 0x4d4) != 0) ||
               (*(int *)(lVar8 + 0x1a0) != 0)) && (iVar6 = Py_MakePendingCalls(), iVar6 < 0)) &&
             (*(longlong *)(param_1 + 0x60) != 0)) goto LAB_18004e401;
          if (*(int *)(lVar8 + 0x174) != 0) {
            PyEval_SaveThread();
            PyEval_AcquireThread(param_1);
          }
        } while (*(longlong *)(param_1 + 0x80) == 0);
        plVar11 = *(longlong **)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x80) = 0;
        uVar18 = FUN_180373490(param_1);
        *(undefined8 *)(param_1 + 0x60) = uVar18;
        if (((plVar11 != (longlong *)0x0) && (-1 < (int)*plVar11)) &&
           (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
          (**(code **)(plVar11[1] + 0x30))(plVar11);
        }
LAB_18004e401:
        plVar11 = *(longlong **)(param_1 + 0x60);
        uVar18 = 0x91;
        *(undefined8 *)(param_1 + 0x60) = 0;
LAB_18004e549:
        local_68 = local_58;
        if ((-1 < (int)*local_58) && (*local_58 = *local_58 + -1, *local_58 == 0)) {
          (**(code **)(local_58[1] + 0x30))(local_58);
        }
LAB_18004e568:
        if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
          (**(code **)(plVar10[1] + 0x30))();
        }
      }
    }
  }
LAB_18004e592:
  piVar3 = (int *)plVar11[5];
  if (piVar3 == (int *)0x0) {
    pcVar15 = (code *)FUN_18038e2a0(plVar9,uVar18);
    plVar10 = (longlong *)plVar11[5];
    pcVar16 = pcVar15;
    if (pcVar15 == _Py_NoneStruct_exref) {
      pcVar16 = (code *)0x0;
    }
    else {
LAB_18004e83a:
      pcVar15 = pcVar16;
      if ((pcVar16 != (code *)0x0) && (*(int *)pcVar16 + 1 != 0)) {
        *(int *)pcVar16 = *(int *)pcVar16 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar3 + 6) == plVar9) goto LAB_18004e883;
    pcVar15 = (code *)FUN_18038e2a0(plVar9,uVar18);
    *(int **)(pcVar15 + 0x10) = piVar3;
    if (*piVar3 + 1 != 0) {
      *piVar3 = *piVar3 + 1;
    }
    plVar10 = (longlong *)plVar11[5];
    pcVar16 = pcVar15;
    if (pcVar15 != _Py_NoneStruct_exref) goto LAB_18004e83a;
    pcVar16 = (code *)0x0;
  }
  plVar11[5] = (longlong)pcVar16;
  if (((plVar10 != (longlong *)0x0) && (-1 < (int)*plVar10)) &&
     (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
    (**(code **)(plVar10[1] + 0x30))();
  }
  if (((pcVar15 != (code *)0x0) && (-1 < *(int *)pcVar15)) &&
     (*(longlong *)pcVar15 = *(longlong *)pcVar15 + -1, *(longlong *)pcVar15 == 0)) {
    (**(code **)(*(longlong *)(pcVar15 + 8) + 0x30))(pcVar15);
  }
LAB_18004e883:
  FUN_18036d0a0(plVar9,"ooooo",plVar1,plVar7,local_68,local_60,plVar17);
  plVar10 = DAT_1803aec48;
  if (plVar9 == DAT_1803aec48) {
    if ((-1 < (int)*DAT_1803aec48) && (*DAT_1803aec48 = *DAT_1803aec48 + -1, *plVar10 == 0)) {
      (**(code **)(plVar10[1] + 0x30))();
    }
    DAT_1803aec48 = (longlong *)0x0;
  }
  lVar8 = **(longlong **)(param_1 + 0x38);
  plVar10 = *(longlong **)(lVar8 + 0x30);
  **(longlong **)(param_1 + 0x38) = *(longlong *)(lVar8 + 8);
  *(undefined4 *)(plVar10 + 8) = 0xffffffff;
  plVar9 = (longlong *)plVar10[2];
  if (((plVar9 != (longlong *)0x0) && (plVar10[2] = 0, -1 < (int)*plVar9)) &&
     (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
    (**(code **)(plVar9[1] + 0x30))();
  }
  if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
    (**(code **)(plVar10[1] + 0x30))(plVar10);
  }
  plVar10[10] = 0;
  if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
    (**(code **)(plVar7[1] + 0x30))(plVar7);
  }
  if (((local_68 != (longlong *)0x0) && (-1 < (int)*local_68)) &&
     (*local_68 = *local_68 + -1, *local_68 == 0)) {
    (**(code **)(local_68[1] + 0x30))(local_68);
  }
  if (((local_60 != (longlong *)0x0) && (-1 < (int)*local_60)) &&
     (*local_60 = *local_60 + -1, *local_60 == 0)) {
    (**(code **)(local_60[1] + 0x30))(local_60);
  }
  if (((plVar17 != (longlong *)0x0) && (-1 < (int)*plVar17)) &&
     (*plVar17 = *plVar17 + -1, *plVar17 == 0)) {
    (**(code **)(plVar17[1] + 0x30))(plVar17);
  }
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong **)(param_1 + 0x60) = plVar11;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return (longlong *)0x0;
}

