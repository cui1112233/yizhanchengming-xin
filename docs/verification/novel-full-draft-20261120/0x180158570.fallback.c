/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x180158570 */

longlong FUN_180158570(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  undefined8 *puVar2;
  int *piVar3;
  longlong *plVar4;
  char cVar5;
  int iVar6;
  longlong lVar7;
  longlong *plVar8;
  longlong *plVar9;
  longlong *plVar10;
  longlong *plVar11;
  longlong *plVar12;
  code *pcVar13;
  code *pcVar14;
  longlong lVar15;
  undefined8 uVar16;
  longlong *plVar17;
  longlong *plVar18;
  longlong lVar19;
  longlong *local_res18;
  longlong *local_res20;
  longlong *local_78;
  longlong *local_58 [2];
  longlong *local_48;
  longlong *local_40;

  plVar10 = DAT_1803b6ae0;
  plVar1 = (longlong *)*param_3;
  plVar12 = (longlong *)0x0;
  plVar18 = (longlong *)0x0;
  plVar17 = (longlong *)0x0;
  local_res18 = (longlong *)0x0;
  local_78 = (longlong *)0x0;
  local_58[0] = DAT_180391d20;
  if (DAT_1803b6ae0 == (longlong *)0x0) {
LAB_1801585f4:
    DAT_1803b6ae0 = (longlong *)FUN_18036cd10(param_1,DAT_1803b6ba0,DAT_1803ca758,0x20);
  }
  else if ((1 < *DAT_1803b6ae0) || (DAT_1803b6ae0[2] != 0)) {
    if ((DAT_1803b6ae0 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b6ae0 && (*DAT_1803b6ae0 = *DAT_1803b6ae0 + -1, *plVar10 == 0)))) {
      (**(code **)(plVar10[1] + 0x30))();
    }
    goto LAB_1801585f4;
  }
  plVar4 = DAT_1803b6ae0;
  FUN_180001d00(param_1,DAT_1803b6ae0);
  piVar3 = DAT_1803b6f10;
  lVar7 = *(longlong *)(param_1 + 0x10);
  plVar10 = *(longlong **)(lVar7 + 0x41da0);
  if (plVar10 == (longlong *)0x0) {
    plVar10 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,2);
  }
  else {
    *(longlong *)(lVar7 + 0x41da0) = plVar10[3];
    *(int *)(lVar7 + 0x41e3c) = *(int *)(lVar7 + 0x41e3c) + -1;
    *plVar10 = 1;
  }
  plVar10[3] = 0;
  plVar10[4] = 0;
  plVar9 = plVar10 + -2;
  lVar7 = _PyThreadState_GetCurrent();
  lVar7 = *(longlong *)(lVar7 + 0x10);
  lVar15 = *(longlong *)(lVar7 + 0xd0);
  puVar2 = *(undefined8 **)(lVar15 + 8);
  *puVar2 = plVar9;
  plVar10[-1] = plVar10[-1] & 3;
  plVar10[-1] = plVar10[-1] | (ulonglong)puVar2;
  *plVar9 = lVar15;
  *(longlong **)(lVar15 + 8) = plVar9;
  if (*piVar3 + 1 != 0) {
    *piVar3 = *piVar3 + 1;
  }
  plVar10[3] = (longlong)piVar3;
  plVar8 = (longlong *)FUN_18037f0c0(lVar7,plVar1,DAT_1803b6f18);
  plVar9 = plVar12;
  plVar11 = plVar17;
  if (plVar8 == (longlong *)0x0) {
LAB_1801586fb:
    uVar16 = 0x113;
LAB_180158700:
    plVar8 = *(longlong **)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
    plVar11 = plVar12;
    local_78 = plVar12;
    if ((-1 < (int)*plVar10) &&
       (*plVar10 = *plVar10 + -1, plVar11 = plVar17, local_78 = plVar18, *plVar10 == 0)) {
      (**(code **)(plVar10[1] + 0x30))(plVar10);
    }
  }
  else {
    lVar7 = FUN_180375520(param_1,plVar8,DAT_1803b6c50);
    if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
      (**(code **)(plVar8[1] + 0x30))(plVar8);
    }
    if (lVar7 == 0) goto LAB_1801586fb;
    plVar10[4] = lVar7;
    lVar7 = PyUnicode_Join(DAT_1803b6c50,plVar10);
    if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
      (**(code **)(plVar10[1] + 0x30))(plVar10);
    }
    if (lVar7 != 0) {
      plVar9 = (longlong *)FUN_180372160(param_1,1);
      *(longlong *)plVar9[3] = lVar7;
      local_res20 = plVar9;
      plVar10 = (longlong *)FUN_18037f0c0();
      if (plVar10 == (longlong *)0x0) {
        plVar8 = *(longlong **)(param_1 + 0x60);
        uVar16 = 0x114;
        *(undefined8 *)(param_1 + 0x60) = 0;
        local_78 = plVar18;
        goto LAB_1801591c5;
      }
      iVar6 = FUN_180001510(plVar10);
      if (iVar6 == -1) {
        plVar8 = *(longlong **)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
        if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
          (**(code **)(plVar10[1] + 0x30))(plVar10);
        }
        uVar16 = 0x114;
        local_78 = plVar18;
        goto LAB_1801591c5;
      }
      if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
        (**(code **)(plVar10[1] + 0x30))(plVar10);
      }
      piVar3 = DAT_1803b6f28;
      if (iVar6 == 0) {
LAB_18015894c:
        plVar10 = (longlong *)FUN_18037f0c0();
        if (plVar10 == (longlong *)0x0) {
          plVar8 = *(longlong **)(param_1 + 0x60);
          uVar16 = 0x116;
          *(undefined8 *)(param_1 + 0x60) = 0;
          local_78 = plVar18;
        }
        else {
          plVar17 = (longlong *)FUN_18000d280();
          if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
            (**(code **)(plVar10[1] + 0x30))(plVar10);
          }
          if (plVar17 == (longlong *)0x0) {
            plVar8 = *(longlong **)(param_1 + 0x60);
            uVar16 = 0x116;
            *(undefined8 *)(param_1 + 0x60) = 0;
            local_78 = plVar18;
          }
          else {
            uVar16 = 0x117;
            do {
              pcVar13 = *(code **)(plVar17[1] + 0xe0);
              if (pcVar13 == (code *)0x0) {
                PyErr_Format(*(undefined8 *)PyExc_TypeError_exref,"\'%s\' object is not an iterator"
                             ,*(undefined8 *)(plVar17[1] + 0x18));
                plVar11 = plVar12;
LAB_180158fb9:
                plVar8 = *(longlong **)(param_1 + 0x60);
                if (plVar8 != (longlong *)0x0) {
                  plVar10 = (longlong *)plVar8[1];
                  if ((*(uint *)((longlong *)plVar8[1] + 0x15) & 0x40000000) == 0) {
                    plVar10 = plVar8;
                  }
                  if ((plVar10 != *(longlong **)PyExc_StopIteration_exref) &&
                     (((-1 < *(int *)(plVar10[1] + 0xa8) ||
                       ((*(uint *)(plVar10 + 0x15) & 0x40000000) == 0)) ||
                      (cVar5 = FUN_180382640(), cVar5 != '\x01')))) {
                    *(undefined8 *)(param_1 + 0x60) = 0;
                    uVar16 = 0x116;
                    if (local_78 != (longlong *)0x0) goto LAB_180158ea0;
                    goto LAB_180158ebe;
                  }
                  *(undefined8 *)(param_1 + 0x60) = 0;
                  if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                    (**(code **)(plVar8[1] + 0x30))(plVar8);
                  }
                }
                if (((local_78 != (longlong *)0x0) && (-1 < (int)*local_78)) &&
                   (*local_78 = *local_78 + -1, *local_78 == 0)) {
                  (**(code **)(local_78[1] + 0x30))();
                }
                if ((-1 < (int)*plVar17) && (*plVar17 = *plVar17 + -1, *plVar17 == 0)) {
                  (**(code **)(plVar17[1] + 0x30))();
                }
                piVar3 = DAT_1803b6f50;
                if (plVar9 == (longlong *)0x0) {
                  plVar10 = (longlong *)
                            PyUnicode_FromFormat
                                      ("cannot access local variable \'%U\' where it is not associated with a value"
                                       ,DAT_1803b6f40);
                  plVar8 = (longlong *)
                           FUN_180376770(param_1,*(undefined8 *)PyExc_UnboundLocalError_exref,
                                         plVar10);
                  if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                    (**(code **)(plVar10[1] + 0x30))(plVar10);
                  }
                  FUN_18000d1d0(param_1,plVar8);
                }
                else {
                  plVar10 = (longlong *)FUN_18000d280();
                  if (plVar10 == (longlong *)0x0) {
                    plVar8 = *(longlong **)(param_1 + 0x60);
                    *(undefined8 *)(param_1 + 0x60) = 0;
                  }
                  else {
                    local_res18 = (longlong *)FUN_18036d750(plVar10);
                    plVar12 = (longlong *)
                              FUN_180367890(FUN_180159410,DAT_1803ca758,DAT_1803b6f00,DAT_1803b6f58,
                                            DAT_1803b6b38,&local_res18,1,0x48);
                    if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                      (**(code **)(plVar10[1] + 0x30))(plVar10);
                    }
                    lVar7 = PyUnicode_Join(piVar3,plVar12);
                    if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                      (**(code **)(plVar12[1] + 0x30))(plVar12);
                    }
                    if (lVar7 != 0) {
                      FUN_180001d80(param_1);
                      if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                        (**(code **)(plVar9[1] + 0x30))(plVar9);
                      }
                      if (((local_78 != (longlong *)0x0) && (-1 < (int)*local_78)) &&
                         (*local_78 = *local_78 + -1, *local_78 == 0)) {
                        (**(code **)(local_78[1] + 0x30))();
                      }
                      if (((plVar11 != (longlong *)0x0) && (-1 < (int)*plVar11)) &&
                         (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                        (**(code **)(plVar11[1] + 0x30))(plVar11);
                      }
                      if ((int)*plVar1 < 0) {
                        return lVar7;
                      }
                      *plVar1 = *plVar1 + -1;
                      if (*plVar1 != 0) {
                        return lVar7;
                      }
                      (**(code **)(plVar1[1] + 0x30))();
                      return lVar7;
                    }
                    plVar8 = *(longlong **)(param_1 + 0x60);
                    *(undefined8 *)(param_1 + 0x60) = 0;
                  }
                }
                uVar16 = 0x119;
                goto LAB_1801591c5;
              }
              plVar10 = (longlong *)(*pcVar13)();
              plVar11 = local_res18;
              if (plVar10 == (longlong *)0x0) goto LAB_180158fb9;
              if (((local_78 != (longlong *)0x0) && (-1 < (int)*local_78)) &&
                 (*local_78 = *local_78 + -1, *local_78 == 0)) {
                (**(code **)(local_78[1] + 0x30))(local_78);
              }
              iVar6 = (int)*plVar10 + 1;
              if (iVar6 != 0) {
                *(int *)plVar10 = iVar6;
              }
              if (((local_78 != (longlong *)0x0) && (-1 < (int)*local_78)) &&
                 (*local_78 = *local_78 + -1, *local_78 == 0)) {
                (**(code **)(local_78[1] + 0x30))();
              }
              lVar7 = *(longlong *)(DAT_1803ca760 + 0x20);
              local_78 = plVar10;
              if (*(char *)(lVar7 + 10) == '\0') {
                plVar18 = (longlong *)FUN_1800015a0(DAT_1803ca760,DAT_1803b6f30);
                if (plVar18 == (longlong *)0x0) {
                  lVar15 = 0;
                }
                else {
                  lVar15 = *plVar18;
                }
LAB_180158b13:
                if (lVar15 == 0) goto LAB_180158b18;
              }
              else {
                iVar6 = *(int *)(lVar7 + 0xc);
                if (iVar6 == 0) {
                  iVar6 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
                  *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar6 + 1;
                  *(int *)(lVar7 + 0xc) = iVar6;
                }
                if (iVar6 != DAT_1803a9904) {
                  DAT_1803a9904 = iVar6;
                  DAT_1803b6b30 =
                       FUN_1803707d0(lVar7,DAT_1803b6f30,*(undefined8 *)(DAT_1803b6f30 + 0x18));
                }
                if (-1 < DAT_1803b6b30) {
                  lVar19 = (1L << (*(byte *)(lVar7 + 9) & 0x3f)) + lVar7;
                  lVar15 = *(longlong *)(lVar19 + 0x28 + DAT_1803b6b30 * 0x10);
                  if (lVar15 != 0) goto LAB_180158b40;
                  DAT_1803b6b30 =
                       FUN_1803707d0(lVar7,DAT_1803b6f30,*(undefined8 *)(DAT_1803b6f30 + 0x18));
                  if (-1 < DAT_1803b6b30) {
                    lVar15 = *(longlong *)(lVar19 + 0x28 + DAT_1803b6b30 * 0x10);
                    goto LAB_180158b13;
                  }
                }
LAB_180158b18:
                plVar18 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803b6f30);
                if ((plVar18 == (longlong *)0x0) || (lVar15 = *plVar18, lVar15 == 0)) {
                  FUN_180373340(param_1,local_58,DAT_1803b6f30);
                  plVar8 = local_58[0];
                  plVar11 = local_res18;
                  goto LAB_180158ea0;
                }
              }
LAB_180158b40:
              plVar18 = (longlong *)FUN_18037f0c0();
              if (plVar18 == (longlong *)0x0) {
                plVar8 = *(longlong **)(param_1 + 0x60);
                *(undefined8 *)(param_1 + 0x60) = 0;
                plVar11 = local_res18;
                goto LAB_180158ea0;
              }
              plVar11 = (longlong *)FUN_18037f0c0();
              if (plVar11 == (longlong *)0x0) {
                plVar8 = *(longlong **)(param_1 + 0x60);
                *(undefined8 *)(param_1 + 0x60) = 0;
                plVar11 = local_res18;
                if ((-1 < (int)*plVar18) && (*plVar18 = *plVar18 + -1, *plVar18 == 0)) {
                  (**(code **)(plVar18[1] + 0x30))(plVar18);
                  plVar11 = local_res18;
                }
                goto LAB_180158ea0;
              }
              *(undefined4 *)(plVar4 + 5) = 0x117;
              local_48 = plVar18;
              local_40 = plVar11;
              plVar9 = (longlong *)FUN_180377750(param_1,lVar15,&local_48);
              if ((-1 < (int)*plVar18) && (*plVar18 = *plVar18 + -1, *plVar18 == 0)) {
                (**(code **)(plVar18[1] + 0x30))(plVar18);
              }
              if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                (**(code **)(plVar11[1] + 0x30))(plVar11);
              }
              plVar18 = local_res18;
              if (plVar9 == (longlong *)0x0) {
                plVar8 = *(longlong **)(param_1 + 0x60);
                *(undefined8 *)(param_1 + 0x60) = 0;
                plVar11 = local_res18;
                plVar9 = local_res20;
                goto LAB_180158ea0;
              }
              plVar11 = plVar9;
              if (((local_res18 != (longlong *)0x0) && (-1 < (int)*plVar12)) &&
                 (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                plVar12 = local_res18 + 1;
                local_res18 = plVar9;
                (**(code **)(*plVar12 + 0x30))(plVar18);
                plVar11 = local_res18;
              }
              local_res18 = plVar11;
              piVar3 = DAT_1803b6f48;
              if (local_res20 == (longlong *)0x0) {
                plVar10 = (longlong *)
                          PyUnicode_FromFormat
                                    ("cannot access local variable \'%U\' where it is not associated with a value"
                                     ,DAT_1803b6f40);
                plVar8 = (longlong *)
                         FUN_180376770(param_1,*(undefined8 *)PyExc_UnboundLocalError_exref,plVar10)
                ;
                if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                  (**(code **)(plVar10[1] + 0x30))(plVar10);
                }
                FUN_18000d1d0(param_1,plVar8);
                uVar16 = 0x118;
                plVar11 = local_res18;
                plVar9 = local_res20;
                goto LAB_180158ea0;
              }
              lVar7 = *(longlong *)(param_1 + 0x10);
              plVar12 = *(longlong **)(lVar7 + 0x41db0);
              if (plVar12 == (longlong *)0x0) {
                plVar12 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,4);
              }
              else {
                *(longlong *)(lVar7 + 0x41db0) = plVar12[3];
                *(int *)(lVar7 + 0x41e44) = *(int *)(lVar7 + 0x41e44) + -1;
                *plVar12 = 1;
              }
              plVar18 = plVar12 + -2;
              plVar12[3] = 0;
              plVar12[4] = 0;
              plVar12[5] = 0;
              plVar12[6] = 0;
              lVar7 = _PyThreadState_GetCurrent();
              lVar7 = *(longlong *)(*(longlong *)(lVar7 + 0x10) + 0xd0);
              puVar2 = *(undefined8 **)(lVar7 + 8);
              *puVar2 = plVar18;
              plVar12[-1] = plVar12[-1] & 3;
              plVar12[-1] = plVar12[-1] | (ulonglong)puVar2;
              *plVar18 = lVar7;
              *(longlong **)(lVar7 + 8) = plVar18;
              if (*piVar3 + 1 != 0) {
                *piVar3 = *piVar3 + 1;
              }
              plVar12[3] = (longlong)piVar3;
              lVar7 = FUN_180375520(param_1,plVar9,DAT_1803b6c50);
              if (lVar7 == 0) {
LAB_180158e63:
                plVar8 = *(longlong **)(param_1 + 0x60);
                uVar16 = 0x118;
                *(undefined8 *)(param_1 + 0x60) = 0;
                plVar11 = local_res18;
                plVar9 = local_res20;
                if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                  (**(code **)(plVar12[1] + 0x30))(plVar12);
                  plVar11 = local_res18;
                  plVar9 = local_res20;
                }
                goto LAB_180158ea0;
              }
              plVar12[4] = lVar7;
              piVar3 = DAT_1803b6f50;
              if (*DAT_1803b6f50 + 1 != 0) {
                *DAT_1803b6f50 = *DAT_1803b6f50 + 1;
              }
              plVar12[5] = (longlong)piVar3;
              plVar10 = (longlong *)FUN_18037f0c0(piVar3,plVar10,DAT_1803b6eb0);
              if (plVar10 == (longlong *)0x0) goto LAB_180158e63;
              lVar7 = FUN_180375520(param_1,plVar10,DAT_1803b6c50);
              if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                (**(code **)(plVar10[1] + 0x30))(plVar10);
              }
              if (lVar7 == 0) goto LAB_180158e63;
              plVar12[6] = lVar7;
              lVar7 = PyUnicode_Join(DAT_1803b6c50,plVar12);
              if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                (**(code **)(plVar12[1] + 0x30))(plVar12);
              }
              plVar9 = local_res20;
              if (lVar7 == 0) {
                plVar8 = *(longlong **)(param_1 + 0x60);
                uVar16 = 0x118;
                *(undefined8 *)(param_1 + 0x60) = 0;
                plVar11 = local_res18;
                goto LAB_180158ea0;
              }
              lVar15 = local_res20[2];
              cVar5 = FUN_1803722b0(local_res20,lVar15 + 1);
              if (cVar5 != '\0') {
                *(longlong *)(plVar9[3] + lVar15 * 8) = lVar7;
              }
              lVar7 = *(longlong *)(param_1 + 0x10);
              if ((((*(int *)(*(longlong *)(lVar7 + 0x60) + 0x4d4) != 0) ||
                   (*(int *)(lVar7 + 0x1a0) != 0)) && (iVar6 = Py_MakePendingCalls(), iVar6 < 0)) &&
                 (*(longlong *)(param_1 + 0x60) != 0)) goto LAB_180158e47;
              if (*(int *)(lVar7 + 0x174) != 0) {
                PyEval_SaveThread();
                PyEval_AcquireThread(param_1);
              }
              plVar12 = local_res18;
            } while (*(longlong *)(param_1 + 0x80) == 0);
            plVar10 = *(longlong **)(param_1 + 0x60);
            *(undefined8 *)(param_1 + 0x80) = 0;
            uVar16 = FUN_180373490(param_1);
            *(undefined8 *)(param_1 + 0x60) = uVar16;
            if (((plVar10 != (longlong *)0x0) && (-1 < (int)*plVar10)) &&
               (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
              (**(code **)(plVar10[1] + 0x30))(plVar10);
            }
LAB_180158e47:
            plVar8 = *(longlong **)(param_1 + 0x60);
            uVar16 = 0x116;
            *(undefined8 *)(param_1 + 0x60) = 0;
            plVar11 = local_res18;
LAB_180158ea0:
            if ((-1 < (int)*local_78) && (*local_78 = *local_78 + -1, *local_78 == 0)) {
              (**(code **)(local_78[1] + 0x30))(local_78);
            }
LAB_180158ebe:
            if ((-1 < (int)*plVar17) && (*plVar17 = *plVar17 + -1, *plVar17 == 0)) {
              (**(code **)(plVar17[1] + 0x30))();
            }
          }
        }
        goto LAB_1801591c5;
      }
      lVar7 = *(longlong *)(param_1 + 0x10);
      plVar10 = *(longlong **)(lVar7 + 0x41da0);
      if (plVar10 == (longlong *)0x0) {
        plVar10 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,2);
      }
      else {
        *(longlong *)(lVar7 + 0x41da0) = plVar10[3];
        *(int *)(lVar7 + 0x41e3c) = *(int *)(lVar7 + 0x41e3c) + -1;
        *plVar10 = 1;
      }
      plVar8 = plVar10 + -2;
      plVar10[3] = 0;
      plVar10[4] = 0;
      lVar7 = _PyThreadState_GetCurrent();
      lVar7 = *(longlong *)(lVar7 + 0x10);
      lVar15 = *(longlong *)(lVar7 + 0xd0);
      puVar2 = *(undefined8 **)(lVar15 + 8);
      *puVar2 = plVar8;
      plVar10[-1] = plVar10[-1] & 3;
      plVar10[-1] = plVar10[-1] | (ulonglong)puVar2;
      *plVar8 = lVar15;
      *(longlong **)(lVar15 + 8) = plVar8;
      if (*piVar3 + 1 != 0) {
        *piVar3 = *piVar3 + 1;
      }
      plVar10[3] = (longlong)piVar3;
      plVar8 = (longlong *)FUN_18037f0c0(lVar7,plVar1,DAT_1803b6f20);
      if (plVar8 != (longlong *)0x0) {
        lVar7 = FUN_180375520(param_1,plVar8,DAT_1803b6c50);
        if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
          (**(code **)(plVar8[1] + 0x30))(plVar8);
        }
        if (lVar7 != 0) {
          plVar10[4] = lVar7;
          lVar7 = PyUnicode_Join(DAT_1803b6c50,plVar10);
          if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
            (**(code **)(plVar10[1] + 0x30))(plVar10);
          }
          if (lVar7 == 0) {
            plVar8 = *(longlong **)(param_1 + 0x60);
            uVar16 = 0x115;
            *(undefined8 *)(param_1 + 0x60) = 0;
            local_78 = plVar18;
            goto LAB_1801591c5;
          }
          FUN_180372560(plVar9,lVar7);
          goto LAB_18015894c;
        }
      }
      uVar16 = 0x115;
      goto LAB_180158700;
    }
    plVar8 = *(longlong **)(param_1 + 0x60);
    uVar16 = 0x113;
    *(undefined8 *)(param_1 + 0x60) = 0;
    local_78 = plVar18;
  }
LAB_1801591c5:
  piVar3 = (int *)plVar8[5];
  if (piVar3 == (int *)0x0) {
    pcVar13 = (code *)FUN_18038e2a0(plVar4,uVar16);
    plVar10 = (longlong *)plVar8[5];
    pcVar14 = pcVar13;
    if (pcVar13 == _Py_NoneStruct_exref) {
      pcVar14 = (code *)0x0;
    }
    else {
LAB_180159226:
      pcVar13 = pcVar14;
      if ((pcVar14 != (code *)0x0) && (*(int *)pcVar14 + 1 != 0)) {
        *(int *)pcVar14 = *(int *)pcVar14 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar3 + 6) == plVar4) goto LAB_18015926f;
    pcVar13 = (code *)FUN_18038e2a0(plVar4,uVar16);
    *(int **)(pcVar13 + 0x10) = piVar3;
    if (*piVar3 + 1 != 0) {
      *piVar3 = *piVar3 + 1;
    }
    plVar10 = (longlong *)plVar8[5];
    pcVar14 = pcVar13;
    if (pcVar13 != _Py_NoneStruct_exref) goto LAB_180159226;
    pcVar14 = (code *)0x0;
  }
  plVar8[5] = (longlong)pcVar14;
  if (((plVar10 != (longlong *)0x0) && (-1 < (int)*plVar10)) &&
     (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
    (**(code **)(plVar10[1] + 0x30))();
  }
  if (((pcVar13 != (code *)0x0) && (-1 < *(int *)pcVar13)) &&
     (*(longlong *)pcVar13 = *(longlong *)pcVar13 + -1, *(longlong *)pcVar13 == 0)) {
    (**(code **)(*(longlong *)(pcVar13 + 8) + 0x30))(pcVar13);
  }
LAB_18015926f:
  FUN_18036d0a0(plVar4,&DAT_180391c1c,plVar1,plVar9,local_78,plVar11);
  plVar10 = DAT_1803b6ae0;
  if (plVar4 == DAT_1803b6ae0) {
    if ((-1 < (int)*DAT_1803b6ae0) && (*DAT_1803b6ae0 = *DAT_1803b6ae0 + -1, *plVar10 == 0)) {
      (**(code **)(plVar10[1] + 0x30))();
    }
    DAT_1803b6ae0 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if (((plVar9 != (longlong *)0x0) && (-1 < (int)*plVar9)) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)
     ) {
    (**(code **)(plVar9[1] + 0x30))(plVar9);
  }
  if (((local_78 != (longlong *)0x0) && (-1 < (int)*local_78)) &&
     (*local_78 = *local_78 + -1, *local_78 == 0)) {
    (**(code **)(local_78[1] + 0x30))(local_78);
  }
  if (((plVar11 != (longlong *)0x0) && (-1 < (int)*plVar11)) &&
     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
    (**(code **)(plVar11[1] + 0x30))(plVar11);
  }
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))(plVar1);
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong **)(param_1 + 0x60) = plVar8;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return 0;
}
