/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x1800d4df0 */

code * FUN_1800d4df0(longlong param_1,longlong param_2,undefined8 *param_3)

{
  undefined8 *puVar1;
  int *piVar2;
  undefined8 uVar3;
  longlong *plVar4;
  undefined8 uVar5;
  undefined8 uVar6;
  undefined8 uVar7;
  int iVar8;
  int iVar9;
  longlong *plVar10;
  longlong lVar11;
  code *pcVar12;
  longlong *plVar13;
  longlong *plVar14;
  code *pcVar15;
  code *pcVar16;
  undefined8 uVar17;
  longlong *plVar18;
  longlong *plVar19;
  longlong *plVar20;
  longlong *local_res20;
  code *local_68;
  longlong *local_60;
  longlong *local_58;
  longlong *local_50;
  longlong *plVar21;

  plVar4 = DAT_1803b2a60;
  plVar10 = (longlong *)*param_3;
  plVar13 = (longlong *)0x0;
  plVar19 = (longlong *)0x0;
  plVar21 = (longlong *)0x0;
  plVar20 = (longlong *)0x0;
  local_60 = (longlong *)0x0;
  local_50 = plVar10;
  if (DAT_1803b2a60 == (longlong *)0x0) {
LAB_1800d4e5f:
    DAT_1803b2a60 = (longlong *)FUN_18036cd10(param_1,DAT_1803b2970,DAT_1803ca7a0,0x30);
  }
  else if ((1 < *DAT_1803b2a60) || (DAT_1803b2a60[2] != 0)) {
    if ((DAT_1803b2a60 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b2a60 && (*DAT_1803b2a60 = *DAT_1803b2a60 + -1, *plVar4 == 0)))) {
      (**(code **)(plVar4[1] + 0x30))(plVar4);
    }
    goto LAB_1800d4e5f;
  }
  plVar4 = DAT_1803b2a60;
  FUN_180001d00(param_1,DAT_1803b2a60);
  lVar11 = *(longlong *)(*(longlong *)(param_2 + 0xe8) + 0x10);
  plVar14 = plVar21;
  if (lVar11 == 0) {
    lVar11 = DAT_1803b2be8 + 0x28;
    if (((byte)*(undefined4 *)(DAT_1803b2be8 + 0x20) & 0x60) != 0x60) {
      lVar11 = *(longlong *)(DAT_1803b2be8 + 0x30);
    }
    plVar10 = (longlong *)
              PyUnicode_FromFormat
                        ("cannot access free variable \'%s\' where it is not associated with a value in enclosing scope"
                         ,lVar11);
    lVar11 = FUN_180376770(param_1,*(undefined8 *)PyExc_NameError_exref,plVar10);
    if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
      (**(code **)(plVar10[1] + 0x30))(plVar10);
    }
    FUN_18000d1d0(param_1,lVar11);
    uVar17 = 0x1d5;
    pcVar12 = (code *)0x0;
  }
  else {
    uVar17 = 0x1d5;
    *(undefined4 *)(plVar4 + 5) = 0x1d5;
    pcVar12 = (code *)FUN_18037e6c0(param_1,lVar11,DAT_1803b2bf0,plVar10);
    if (pcVar12 == (code *)0x0) {
      lVar11 = *(longlong *)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      pcVar12 = (code *)0x0;
    }
    else {
      iVar8 = FUN_1803836d0(plVar10,DAT_1803b3150);
      uVar7 = DAT_1803ca798;
      uVar6 = DAT_1803b3160;
      uVar5 = DAT_1803b3158;
      uVar3 = DAT_1803ab250;
      pcVar16 = _Py_NoneStruct_exref;
      if (iVar8 == -1) {
        lVar11 = *(longlong *)(param_1 + 0x60);
        uVar17 = 0x1d6;
        *(undefined8 *)(param_1 + 0x60) = 0;
      }
      else {
        plVar18 = (longlong *)0x0;
        pcVar15 = pcVar12;
        if (iVar8 != 1) {
LAB_1800d58e0:
          FUN_180001d80(param_1);
          iVar8 = *(int *)pcVar15 + 1;
          iVar9 = *(int *)pcVar15;
          if (iVar8 != 0) {
            *(int *)pcVar15 = iVar8;
            iVar9 = iVar8;
          }
          if ((-1 < iVar9) &&
             (*(longlong *)pcVar15 = *(longlong *)pcVar15 + -1, *(longlong *)pcVar15 == 0)) {
            (**(code **)(*(longlong *)(pcVar15 + 8) + 0x30))(pcVar15);
          }
          if (((plVar18 != (longlong *)0x0) && (-1 < (int)*plVar18)) &&
             (*plVar18 = *plVar18 + -1, *plVar18 == 0)) {
            (**(code **)(plVar18[1] + 0x30))(plVar18);
          }
          if (((plVar13 != (longlong *)0x0) && (-1 < (int)*plVar13)) &&
             (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
            (**(code **)(plVar13[1] + 0x30))(plVar13);
          }
          if (((plVar14 != (longlong *)0x0) && (-1 < (int)*plVar14)) &&
             (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
            (**(code **)(plVar14[1] + 0x30))(plVar14);
          }
          if ((int)*plVar10 < 0) {
            return pcVar15;
          }
          *plVar10 = *plVar10 + -1;
          if (*plVar10 != 0) {
            return pcVar15;
          }
          (**(code **)(plVar10[1] + 0x30))(plVar10);
          return pcVar15;
        }
        uVar17 = 0x1d7;
        *(undefined4 *)(plVar4 + 5) = 0x1d7;
        plVar10 = (longlong *)FUN_1803717e0(param_1,uVar5,uVar7,pcVar16,uVar6,uVar3);
        plVar13 = plVar19;
        plVar14 = plVar20;
        if (plVar10 == (longlong *)0x0) {
          lVar11 = *(longlong *)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
        }
        else {
          if (((code *)plVar10[1] == PyModule_Type_exref) ||
             (iVar8 = PyType_IsSubtype(), iVar8 != 0)) {
            plVar20 = (longlong *)
                      FUN_180371d80(param_1,plVar10,DAT_1803ca798,DAT_1803b3168,DAT_1803ab250);
          }
          else {
            plVar20 = (longlong *)FUN_180371890(param_1,plVar10,DAT_1803b3168);
          }
          local_58 = plVar20;
          if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
            (**(code **)(plVar10[1] + 0x30))(plVar10);
          }
          uVar7 = DAT_1803ca798;
          uVar6 = DAT_1803b3180;
          uVar5 = DAT_1803b3178;
          uVar3 = DAT_1803b3170;
          pcVar16 = _Py_NoneStruct_exref;
          if (plVar20 == (longlong *)0x0) {
            lVar11 = *(longlong *)(param_1 + 0x60);
            *(undefined8 *)(param_1 + 0x60) = 0;
          }
          else {
            uVar17 = 0x1d8;
            *(undefined4 *)(plVar4 + 5) = 0x1d8;
            local_60 = plVar20;
            plVar10 = (longlong *)FUN_1803717e0(param_1,uVar3,uVar7,pcVar16,uVar5,uVar6);
            if (plVar10 == (longlong *)0x0) {
              lVar11 = *(longlong *)(param_1 + 0x60);
              *(undefined8 *)(param_1 + 0x60) = 0;
            }
            else {
              if (((code *)plVar10[1] == PyModule_Type_exref) ||
                 (iVar8 = PyType_IsSubtype(), iVar8 != 0)) {
                plVar13 = (longlong *)
                          FUN_180371d80(param_1,plVar10,DAT_1803ca798,DAT_1803b3188,DAT_1803ab250);
              }
              else {
                plVar13 = (longlong *)FUN_180371890(param_1,plVar10,DAT_1803b3188);
              }
              if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                (**(code **)(plVar10[1] + 0x30))(plVar10);
              }
              if (plVar13 == (longlong *)0x0) {
                lVar11 = *(longlong *)(param_1 + 0x60);
                *(undefined8 *)(param_1 + 0x60) = 0;
                plVar13 = plVar19;
              }
              else {
                local_res20 = (longlong *)FUN_180370ec0(param_1);
                plVar10 = (longlong *)FUN_180370ec0(param_1);
                iVar8 = (int)*plVar10 + 1;
                if (iVar8 != 0) {
                  *(int *)plVar10 = iVar8;
                }
                uVar17 = FUN_1800c7460();
                PyDict_SetItem(plVar10,DAT_1803b3190,uVar17);
                PyDict_SetItem(plVar10,DAT_1803b31a0,DAT_1803b3198);
                PyDict_SetItem(plVar10,DAT_1803b31b0,DAT_1803b31a8);
                uVar17 = FUN_18036fb80(param_1,DAT_1803b31b8);
                plVar19 = (longlong *)
                          FUN_1803609a0(FUN_1800d59b0,DAT_1803b31c0,DAT_1803b31c8,DAT_1803b2ab0,0,0,
                                        uVar17,DAT_1803ca7a0,0,0,0);
                PyDict_SetItem(plVar10,DAT_1803b31c0,plVar19);
                if ((-1 < (int)*plVar19) && (*plVar19 = *plVar19 + -1, *plVar19 == 0)) {
                  (**(code **)(plVar19[1] + 0x30))(plVar19);
                }
                local_68 = *(code **)(param_2 + 0xe0);
                if (*(int *)local_68 + 1 != 0) {
                  *(int *)local_68 = *(int *)local_68 + 1;
                }
                plVar19 = (longlong *)
                          FUN_1803609a0(FUN_1800d5c20,DAT_1803b31d0,DAT_1803b31d8,DAT_1803b28f8,0,0,
                                        0,DAT_1803ca7a0,0,&local_68,1);
                PyDict_SetItem(plVar10,DAT_1803b31d0,plVar19);
                if ((-1 < (int)*plVar19) && (*plVar19 = *plVar19 + -1, *plVar19 == 0)) {
                  (**(code **)(plVar19[1] + 0x30))(plVar19);
                }
                plVar19 = (longlong *)
                          FUN_1803609a0(FUN_1800d6450,DAT_1803b2c00,DAT_1803b31e0,DAT_1803b2908,0,0,
                                        0,DAT_1803ca7a0,0,0,0);
                PyDict_SetItem(plVar10,DAT_1803b2c00,plVar19);
                if ((-1 < (int)*plVar19) && (*plVar19 = *plVar19 + -1, *plVar19 == 0)) {
                  (**(code **)(plVar19[1] + 0x30))(plVar19);
                }
                plVar19 = (longlong *)
                          FUN_1803609a0(FUN_1800d6820,DAT_1803b31e8,DAT_1803b31f0,DAT_1803b28f0,0,0,
                                        0,DAT_1803ca7a0,0,0,0);
                PyDict_SetItem(plVar10,DAT_1803b31e8,plVar19);
                if ((-1 < (int)*plVar19) && (*plVar19 = *plVar19 + -1, *plVar19 == 0)) {
                  (**(code **)(plVar19[1] + 0x30))(plVar19);
                }
                plVar19 = (longlong *)
                          FUN_1803609a0(FUN_1800d6bf0,DAT_1803b31f8,DAT_1803b3200,DAT_1803b2ac8,0,0,
                                        0,DAT_1803ca7a0,0,0,0);
                PyDict_SetItem(plVar10,DAT_1803b31f8,plVar19);
                if ((-1 < (int)*plVar19) && (*plVar19 = *plVar19 + -1, *plVar19 == 0)) {
                  (**(code **)(plVar19[1] + 0x30))(plVar19);
                }
                piVar2 = DAT_1803b3208;
                lVar11 = *(longlong *)(param_1 + 0x10);
                local_68 = PyType_Type_exref;
                plVar19 = *(longlong **)(lVar11 + 0x41da8);
                if (plVar19 == (longlong *)0x0) {
                  plVar19 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,3);
                }
                else {
                  *(longlong *)(lVar11 + 0x41da8) = plVar19[3];
                  *(int *)(lVar11 + 0x41e40) = *(int *)(lVar11 + 0x41e40) + -1;
                  *plVar19 = 1;
                }
                plVar20 = plVar19 + -2;
                plVar19[3] = 0;
                plVar19[4] = 0;
                plVar19[5] = 0;
                lVar11 = _PyThreadState_GetCurrent();
                lVar11 = *(longlong *)(*(longlong *)(lVar11 + 0x10) + 0xd0);
                puVar1 = *(undefined8 **)(lVar11 + 8);
                *puVar1 = plVar20;
                plVar19[-1] = plVar19[-1] & 3;
                plVar19[-1] = plVar19[-1] | (ulonglong)puVar1;
                *plVar20 = lVar11;
                *(longlong **)(lVar11 + 8) = plVar20;
                if (*piVar2 + 1 != 0) {
                  *piVar2 = *piVar2 + 1;
                }
                plVar19[3] = (longlong)piVar2;
                piVar2 = DAT_1803ab240;
                if (*DAT_1803ab240 + 1 != 0) {
                  *DAT_1803ab240 = *DAT_1803ab240 + 1;
                }
                plVar19[4] = (longlong)piVar2;
                iVar8 = (int)*plVar10 + 1;
                if (iVar8 != 0) {
                  *(int *)plVar10 = iVar8;
                }
                uVar17 = 0x1da;
                plVar19[5] = (longlong)plVar10;
                *(undefined4 *)(plVar4 + 5) = 0x1da;
                plVar14 = (longlong *)FUN_180001790(param_1,local_68,plVar19,local_res20);
                if ((-1 < (int)*plVar19) && (*plVar19 = *plVar19 + -1, *plVar19 == 0)) {
                  (**(code **)(plVar19[1] + 0x30))(plVar19);
                }
                if (plVar14 == (longlong *)0x0) {
                  lVar11 = *(longlong *)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                  if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                    (**(code **)(plVar10[1] + 0x30))(plVar10);
                  }
                  if ((-1 < (int)*local_res20) &&
                     (*local_res20 = *local_res20 + -1, *local_res20 == 0)) {
                    (**(code **)(local_res20[1] + 0x30))();
                  }
                  plVar14 = plVar21;
                  if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                    (**(code **)(plVar10[1] + 0x30))(plVar10);
                  }
                }
                else {
                  iVar8 = (int)*plVar14 + 1;
                  if (iVar8 != 0) {
                    *(int *)plVar14 = iVar8;
                  }
                  if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                    (**(code **)(plVar10[1] + 0x30))(plVar10);
                  }
                  if ((-1 < (int)*plVar14) && (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                    (**(code **)(plVar14[1] + 0x30))(plVar14);
                  }
                  if ((-1 < (int)*local_res20) &&
                     (*local_res20 = *local_res20 + -1, *local_res20 == 0)) {
                    (**(code **)(local_res20[1] + 0x30))();
                  }
                  if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                    (**(code **)(plVar10[1] + 0x30))(plVar10);
                  }
                  plVar10 = (longlong *)FUN_18037f0c0();
                  uVar17 = 0x1f2;
                  if (plVar10 != (longlong *)0x0) {
                    *(undefined4 *)(plVar4 + 5) = 0x1f2;
                    plVar19 = (longlong *)FUN_180376770(param_1,plVar14,plVar10);
                    if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                      (**(code **)(plVar10[1] + 0x30))(plVar10);
                    }
                    if (plVar19 != (longlong *)0x0) {
                      *(undefined4 *)(plVar4 + 5) = 0x1f2;
                      local_res20 = plVar19;
                      local_68 = pcVar12;
                      pcVar15 = (code *)FUN_18037d030(param_1,local_58,&local_68,&local_res20,
                                                      DAT_1803b3218);
                      if ((-1 < (int)*plVar19) && (*plVar19 = *plVar19 + -1, *plVar19 == 0)) {
                        (**(code **)(plVar19[1] + 0x30))(plVar19);
                      }
                      if (pcVar15 != (code *)0x0) {
                        plVar10 = local_50;
                        plVar18 = local_60;
                        if ((-1 < *(int *)pcVar12) &&
                           (*(longlong *)pcVar12 = *(longlong *)pcVar12 + -1,
                           *(longlong *)pcVar12 == 0)) {
                          (**(code **)(*(longlong *)(pcVar12 + 8) + 0x30))();
                          plVar10 = local_50;
                          plVar18 = local_60;
                        }
                        goto LAB_1800d58e0;
                      }
                    }
                  }
                  lVar11 = *(longlong *)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                }
              }
            }
          }
        }
      }
    }
  }
  piVar2 = *(int **)(lVar11 + 0x28);
  if (piVar2 == (int *)0x0) {
    pcVar16 = (code *)FUN_18038e2a0(plVar4,uVar17);
    plVar10 = *(longlong **)(lVar11 + 0x28);
    if (pcVar16 == _Py_NoneStruct_exref) {
      pcVar15 = (code *)0x0;
    }
    else {
      pcVar15 = pcVar16;
      if ((pcVar16 != (code *)0x0) && (*(int *)pcVar16 + 1 != 0)) {
        *(int *)pcVar16 = *(int *)pcVar16 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar2 + 6) == plVar4) goto LAB_1800d577e;
    pcVar16 = (code *)FUN_18038e2a0(plVar4,uVar17);
    *(int **)(pcVar16 + 0x10) = piVar2;
    if (*piVar2 + 1 != 0) {
      *piVar2 = *piVar2 + 1;
    }
    plVar10 = *(longlong **)(lVar11 + 0x28);
    if (pcVar16 == _Py_NoneStruct_exref) {
      pcVar15 = (code *)0x0;
    }
    else {
      pcVar15 = pcVar16;
      if ((pcVar16 != (code *)0x0) && (*(int *)pcVar16 + 1 != 0)) {
        *(int *)pcVar16 = *(int *)pcVar16 + 1;
      }
    }
  }
  *(code **)(lVar11 + 0x28) = pcVar15;
  if (((plVar10 != (longlong *)0x0) && (-1 < (int)*plVar10)) &&
     (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
    (**(code **)(plVar10[1] + 0x30))();
  }
  if (((pcVar16 != (code *)0x0) && (-1 < *(int *)pcVar16)) &&
     (*(longlong *)pcVar16 = *(longlong *)pcVar16 + -1, *(longlong *)pcVar16 == 0)) {
    (**(code **)(*(longlong *)(pcVar16 + 8) + 0x30))(pcVar16);
  }
LAB_1800d577e:
  plVar20 = local_50;
  plVar19 = local_60;
  FUN_18036d0a0(plVar4,"oooooc",local_50,pcVar12,local_60,plVar13,plVar14,
                *(undefined8 *)(param_2 + 0xe8));
  plVar10 = DAT_1803b2a60;
  if (plVar4 == DAT_1803b2a60) {
    if ((-1 < (int)*DAT_1803b2a60) && (*DAT_1803b2a60 = *DAT_1803b2a60 + -1, *plVar10 == 0)) {
      (**(code **)(plVar10[1] + 0x30))();
    }
    DAT_1803b2a60 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if (((pcVar12 != (code *)0x0) && (-1 < *(int *)pcVar12)) &&
     (*(longlong *)pcVar12 = *(longlong *)pcVar12 + -1, *(longlong *)pcVar12 == 0)) {
    (**(code **)(*(longlong *)(pcVar12 + 8) + 0x30))(pcVar12);
  }
  if (((plVar19 != (longlong *)0x0) && (-1 < (int)*plVar19)) &&
     (*plVar19 = *plVar19 + -1, *plVar19 == 0)) {
    (**(code **)(plVar19[1] + 0x30))(plVar19);
  }
  if (((plVar13 != (longlong *)0x0) && (-1 < (int)*plVar13)) &&
     (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
    (**(code **)(plVar13[1] + 0x30))(plVar13);
  }
  if (((plVar14 != (longlong *)0x0) && (-1 < (int)*plVar14)) &&
     (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
    (**(code **)(plVar14[1] + 0x30))(plVar14);
  }
  if ((-1 < (int)*plVar20) && (*plVar20 = *plVar20 + -1, *plVar20 == 0)) {
    (**(code **)(plVar20[1] + 0x30))(plVar20);
  }
  plVar10 = *(longlong **)(param_1 + 0x60);
  *(longlong *)(param_1 + 0x60) = lVar11;
  if (((plVar10 != (longlong *)0x0) && (-1 < (int)*plVar10)) &&
     (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
    (**(code **)(plVar10[1] + 0x30))();
  }
  return (code *)0x0;
}
