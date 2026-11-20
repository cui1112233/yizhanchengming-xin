/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x18014f860 */

int * FUN_18014f860(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  longlong *plVar2;
  undefined8 *puVar3;
  longlong *plVar4;
  int iVar5;
  longlong lVar6;
  longlong *plVar7;
  longlong *plVar8;
  code *pcVar9;
  code *pcVar10;
  longlong *plVar11;
  int *piVar12;
  longlong *plVar13;
  undefined8 uVar14;
  longlong *local_res20;
  longlong *local_68 [2];
  longlong *local_58;
  longlong *local_50;
  longlong *local_48;

  plVar4 = DAT_1803b6bf8;
  plVar1 = (longlong *)*param_3;
  plVar13 = (longlong *)0x0;
  plVar2 = (longlong *)param_3[1];
  local_58 = (longlong *)0x0;
  local_res20 = (longlong *)0x0;
  local_68[0] = DAT_180391d20;
  if (DAT_1803b6bf8 == (longlong *)0x0) {
LAB_18014f8d7:
    DAT_1803b6bf8 = (longlong *)FUN_18036cd10(param_1,DAT_1803b6b90,DAT_1803ca758,0x20);
  }
  else if ((1 < *DAT_1803b6bf8) || (DAT_1803b6bf8[2] != 0)) {
    if ((DAT_1803b6bf8 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b6bf8 && (*DAT_1803b6bf8 = *DAT_1803b6bf8 + -1, *plVar4 == 0)))) {
      (**(code **)(plVar4[1] + 0x30))(plVar4);
    }
    goto LAB_18014f8d7;
  }
  plVar4 = DAT_1803b6bf8;
  FUN_180001d00(param_1,DAT_1803b6bf8);
  lVar6 = FUN_18014c500(param_1);
  if (lVar6 == 0) {
    FUN_180373340(param_1,local_68,DAT_1803b6c30);
    uVar14 = 0x5e;
    plVar8 = local_68[0];
    local_res20 = (longlong *)0x0;
  }
  else {
    plVar7 = (longlong *)FUN_18037f0c0();
    piVar12 = DAT_1803b6ce0;
    if (plVar7 == (longlong *)0x0) {
      plVar8 = *(longlong **)(param_1 + 0x60);
      uVar14 = 0x5e;
      *(undefined8 *)(param_1 + 0x60) = 0;
      local_res20 = (longlong *)0x0;
    }
    else {
      lVar6 = *(longlong *)(param_1 + 0x10);
      plVar11 = *(longlong **)(lVar6 + 0x41db8);
      if (plVar11 == (longlong *)0x0) {
        plVar11 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,5);
      }
      else {
        *(longlong *)(lVar6 + 0x41db8) = plVar11[3];
        *(int *)(lVar6 + 0x41e48) = *(int *)(lVar6 + 0x41e48) + -1;
        *plVar11 = 1;
      }
      plVar11[3] = 0;
      plVar11[4] = 0;
      plVar11[5] = 0;
      plVar11[6] = 0;
      plVar11[7] = 0;
      plVar8 = plVar11 + -2;
      lVar6 = _PyThreadState_GetCurrent();
      lVar6 = *(longlong *)(*(longlong *)(lVar6 + 0x10) + 0xd0);
      puVar3 = *(undefined8 **)(lVar6 + 8);
      *puVar3 = plVar8;
      plVar11[-1] = plVar11[-1] & 3;
      plVar11[-1] = plVar11[-1] | (ulonglong)puVar3;
      *plVar8 = lVar6;
      *(longlong **)(lVar6 + 8) = plVar8;
      if (*piVar12 + 1 != 0) {
        *piVar12 = *piVar12 + 1;
      }
      plVar11[3] = (longlong)piVar12;
      lVar6 = FUN_180375520(param_1,plVar2,DAT_1803b6c50);
      if (lVar6 != 0) {
        plVar11[4] = lVar6;
        piVar12 = DAT_1803b6cf0;
        if (*DAT_1803b6cf0 + 1 != 0) {
          *DAT_1803b6cf0 = *DAT_1803b6cf0 + 1;
        }
        plVar11[5] = (longlong)piVar12;
        plVar8 = (longlong *)FUN_180380b80(plVar2,DAT_1803ab258);
        if (plVar8 != (longlong *)0x0) {
          lVar6 = FUN_180375520(param_1,plVar8,DAT_1803b6c50);
          if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
            (**(code **)(plVar8[1] + 0x30))(plVar8);
          }
          if (lVar6 != 0) {
            plVar11[6] = lVar6;
            piVar12 = DAT_1803b6cf8;
            if (*DAT_1803b6cf8 + 1 != 0) {
              *DAT_1803b6cf8 = *DAT_1803b6cf8 + 1;
            }
            plVar11[7] = (longlong)piVar12;
            plVar13 = (longlong *)PyUnicode_Join(DAT_1803b6c50,plVar11);
            if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
              (**(code **)(plVar11[1] + 0x30))(plVar11);
            }
            if (plVar13 == (longlong *)0x0) {
              plVar8 = *(longlong **)(param_1 + 0x60);
              *(undefined8 *)(param_1 + 0x60) = 0;
              if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                (**(code **)(plVar7[1] + 0x30))(plVar7);
              }
              uVar14 = 0x5f;
            }
            else {
              lVar6 = FUN_18014c500(param_1);
              if (lVar6 == 0) {
                FUN_180373340(param_1,local_68,DAT_1803b6c30);
                if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                  (**(code **)(plVar7[1] + 0x30))(plVar7);
                }
                if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                  (**(code **)(plVar13[1] + 0x30))(plVar13);
                }
                uVar14 = 0x60;
                plVar8 = local_68[0];
              }
              else {
                plVar8 = (longlong *)FUN_18037f0c0();
                if (plVar8 == (longlong *)0x0) {
                  plVar8 = *(longlong **)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                  if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                    (**(code **)(plVar7[1] + 0x30))(plVar7);
                  }
                  if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                    (**(code **)(plVar13[1] + 0x30))(plVar13);
                  }
                  uVar14 = 0x60;
                }
                else {
                  uVar14 = 0x5e;
                  *(undefined4 *)(plVar4 + 5) = 0x5e;
                  local_50 = plVar13;
                  local_48 = plVar8;
                  plVar11 = (longlong *)FUN_180377750(param_1,plVar7,&local_50);
                  local_68[0] = plVar11;
                  if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                    (**(code **)(plVar7[1] + 0x30))(plVar7);
                  }
                  if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                    (**(code **)(plVar13[1] + 0x30))(plVar13);
                  }
                  if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                    (**(code **)(plVar8[1] + 0x30))(plVar8);
                  }
                  if (plVar11 != (longlong *)0x0) {
                    uVar14 = 0x62;
                    *(undefined4 *)(plVar4 + 5) = 0x62;
                    local_58 = plVar11;
                    plVar13 = (longlong *)FUN_18037e6c0(param_1,plVar11,DAT_1803b6ca0,plVar1);
                    if (plVar13 != (longlong *)0x0) {
                      iVar5 = FUN_180001510(plVar13);
                      piVar12 = DAT_1803b6c50;
                      local_res20 = plVar13;
                      if (iVar5 == -1) {
                        plVar8 = *(longlong **)(param_1 + 0x60);
                        uVar14 = 99;
                        *(undefined8 *)(param_1 + 0x60) = 0;
                      }
                      else {
                        if (iVar5 == 0) {
                          if (*DAT_1803b6c50 + 1 != 0) {
                            *DAT_1803b6c50 = *DAT_1803b6c50 + 1;
                          }
LAB_18014fde6:
                          FUN_180001d80(param_1);
                          if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                            (**(code **)(local_58[1] + 0x30))();
                          }
                          if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                            (**(code **)(plVar13[1] + 0x30))(plVar13);
                          }
                          if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
                            (**(code **)(plVar1[1] + 0x30))(plVar1);
                          }
                          if ((int)*plVar2 < 0) {
                            return piVar12;
                          }
                          *plVar2 = *plVar2 + -1;
                          if (*plVar2 != 0) {
                            return piVar12;
                          }
                          (**(code **)(plVar2[1] + 0x30))(plVar2);
                          return piVar12;
                        }
                        uVar14 = 99;
                        *(undefined4 *)(plVar4 + 5) = 99;
                        plVar8 = (longlong *)
                                 FUN_18037e6c0(param_1,plVar13,DAT_1803b6cd0,
                                               *(undefined8 *)(DAT_1803b6cd8 + 0x18));
                        if (plVar8 == (longlong *)0x0) {
                          plVar8 = *(longlong **)(param_1 + 0x60);
                          *(undefined8 *)(param_1 + 0x60) = 0;
                        }
                        else {
                          plVar7 = (longlong *)FUN_18037f0c0();
                          if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                            (**(code **)(plVar8[1] + 0x30))(plVar8);
                          }
                          if (plVar7 != (longlong *)0x0) {
                            *(undefined4 *)(plVar4 + 5) = 99;
                            piVar12 = (int *)FUN_180376210(param_1,plVar7);
                            if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                              (**(code **)(plVar7[1] + 0x30))(plVar7);
                            }
                            plVar11 = local_68[0];
                            if (piVar12 != (int *)0x0) goto LAB_18014fde6;
                            goto LAB_18014fcb0;
                          }
                          plVar8 = *(longlong **)(param_1 + 0x60);
                          *(undefined8 *)(param_1 + 0x60) = 0;
                        }
                      }
                      goto LAB_18014fb3c;
                    }
                  }
LAB_18014fcb0:
                  plVar8 = *(longlong **)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                }
              }
            }
            goto LAB_18014fb3c;
          }
        }
      }
      plVar8 = *(longlong **)(param_1 + 0x60);
      uVar14 = 0x5f;
      *(undefined8 *)(param_1 + 0x60) = 0;
      if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
        (**(code **)(plVar7[1] + 0x30))(plVar7);
      }
      local_res20 = plVar13;
      if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
        (**(code **)(plVar11[1] + 0x30))(plVar11);
      }
    }
  }
LAB_18014fb3c:
  piVar12 = (int *)plVar8[5];
  if (piVar12 == (int *)0x0) {
    pcVar9 = (code *)FUN_18038e2a0(plVar4,uVar14);
    plVar13 = (longlong *)plVar8[5];
    pcVar10 = pcVar9;
    if (pcVar9 == _Py_NoneStruct_exref) {
      pcVar10 = (code *)0x0;
    }
    else {
LAB_18014fea6:
      pcVar9 = pcVar10;
      if ((pcVar10 != (code *)0x0) && (*(int *)pcVar10 + 1 != 0)) {
        *(int *)pcVar10 = *(int *)pcVar10 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar12 + 6) == plVar4) goto LAB_18014feef;
    pcVar9 = (code *)FUN_18038e2a0(plVar4,uVar14);
    *(int **)(pcVar9 + 0x10) = piVar12;
    if (*piVar12 + 1 != 0) {
      *piVar12 = *piVar12 + 1;
    }
    plVar13 = (longlong *)plVar8[5];
    pcVar10 = pcVar9;
    if (pcVar9 != _Py_NoneStruct_exref) goto LAB_18014fea6;
    pcVar10 = (code *)0x0;
  }
  plVar8[5] = (longlong)pcVar10;
  if (((plVar13 != (longlong *)0x0) && (-1 < (int)*plVar13)) &&
     (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
    (**(code **)(plVar13[1] + 0x30))();
  }
  if (((pcVar9 != (code *)0x0) && (-1 < *(int *)pcVar9)) &&
     (*(longlong *)pcVar9 = *(longlong *)pcVar9 + -1, *(longlong *)pcVar9 == 0)) {
    (**(code **)(*(longlong *)(pcVar9 + 8) + 0x30))(pcVar9);
  }
LAB_18014feef:
  plVar7 = local_58;
  FUN_18036d0a0(plVar4,&DAT_180391c1c,plVar1,plVar2,local_58,local_res20);
  plVar13 = DAT_1803b6bf8;
  if (plVar4 == DAT_1803b6bf8) {
    if ((-1 < (int)*DAT_1803b6bf8) && (*DAT_1803b6bf8 = *DAT_1803b6bf8 + -1, *plVar13 == 0)) {
      (**(code **)(plVar13[1] + 0x30))();
    }
    DAT_1803b6bf8 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)
     ) {
    (**(code **)(plVar7[1] + 0x30))(plVar7);
  }
  if (((local_res20 != (longlong *)0x0) && (-1 < (int)*local_res20)) &&
     (*local_res20 = *local_res20 + -1, *local_res20 == 0)) {
    (**(code **)(local_res20[1] + 0x30))(local_res20);
  }
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))(plVar1);
  }
  if ((-1 < (int)*plVar2) && (*plVar2 = *plVar2 + -1, *plVar2 == 0)) {
    (**(code **)(plVar2[1] + 0x30))(plVar2);
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong **)(param_1 + 0x60) = plVar8;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return (int *)0x0;
}
