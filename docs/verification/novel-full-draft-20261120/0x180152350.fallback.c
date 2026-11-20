/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x180152350 */

/* WARNING: Function: __security_check_cookie replaced with injection: security_check_cookie */

longlong * FUN_180152350(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  undefined8 *puVar2;
  int *piVar3;
  longlong lVar4;
  longlong *plVar5;
  undefined8 uVar6;
  int iVar7;
  longlong lVar8;
  longlong *plVar9;
  code *pcVar10;
  code *pcVar11;
  longlong *plVar12;
  longlong *plVar13;
  longlong *plVar14;
  undefined8 uVar15;
  undefined1 auStack_a8 [32];
  longlong *local_88;
  longlong *local_70;
  longlong *local_68;
  longlong local_60;
  longlong *local_58;
  undefined8 local_50;
  longlong *local_48;
  ulonglong local_40;

  plVar5 = DAT_1803b6a20;
  local_40 = DAT_1803a6040 ^ (ulonglong)auStack_a8;
  plVar9 = (longlong *)*param_3;
  plVar1 = (longlong *)param_3[1];
  local_68 = (longlong *)0x0;
  local_60 = DAT_180391d20;
  local_70 = plVar9;
  if (DAT_1803b6a20 == (longlong *)0x0) {
LAB_1801523ce:
    DAT_1803b6a20 = (longlong *)FUN_18036cd10(param_1,DAT_1803b6b88,DAT_1803ca758,0x18);
  }
  else if ((1 < *DAT_1803b6a20) || (DAT_1803b6a20[2] != 0)) {
    if ((DAT_1803b6a20 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b6a20 && (*DAT_1803b6a20 = *DAT_1803b6a20 + -1, *plVar5 == 0)))) {
      (**(code **)(plVar5[1] + 0x30))(plVar5);
    }
    goto LAB_1801523ce;
  }
  plVar5 = DAT_1803b6a20;
  FUN_180001d00(param_1,DAT_1803b6a20);
  lVar8 = FUN_18014c500(param_1);
  plVar13 = (longlong *)0x0;
  if (lVar8 == 0) {
    FUN_180373340(param_1,&local_60,DAT_1803b6c30);
    uVar15 = 0x87;
    lVar8 = local_60;
  }
  else {
    plVar9 = (longlong *)FUN_18037f0c0();
    piVar3 = DAT_1803b6d58;
    if (plVar9 == (longlong *)0x0) {
      lVar8 = *(longlong *)(param_1 + 0x60);
      uVar15 = 0x87;
      *(undefined8 *)(param_1 + 0x60) = 0;
      plVar9 = local_70;
    }
    else {
      lVar8 = *(longlong *)(param_1 + 0x10);
      plVar13 = *(longlong **)(lVar8 + 0x41da8);
      if (plVar13 == (longlong *)0x0) {
        plVar13 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,3);
      }
      else {
        *(longlong *)(lVar8 + 0x41da8) = plVar13[3];
        *(int *)(lVar8 + 0x41e40) = *(int *)(lVar8 + 0x41e40) + -1;
        *plVar13 = 1;
      }
      plVar13[3] = 0;
      plVar12 = plVar13 + -2;
      plVar13[4] = 0;
      plVar13[5] = 0;
      lVar8 = _PyThreadState_GetCurrent();
      lVar8 = *(longlong *)(*(longlong *)(lVar8 + 0x10) + 0xd0);
      puVar2 = *(undefined8 **)(lVar8 + 8);
      *puVar2 = plVar12;
      plVar13[-1] = plVar13[-1] & 3;
      plVar13[-1] = plVar13[-1] | (ulonglong)puVar2;
      *plVar12 = lVar8;
      *(longlong **)(lVar8 + 8) = plVar12;
      if (*piVar3 + 1 != 0) {
        *piVar3 = *piVar3 + 1;
      }
      plVar13[3] = (longlong)piVar3;
      lVar8 = FUN_180375520(param_1,plVar1,DAT_1803b6c50);
      if (lVar8 == 0) {
        lVar8 = *(longlong *)(param_1 + 0x60);
        uVar15 = 0x87;
        *(undefined8 *)(param_1 + 0x60) = 0;
        if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
          (**(code **)(plVar9[1] + 0x30))(plVar9);
        }
        if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
          (**(code **)(plVar13[1] + 0x30))(plVar13);
        }
        plVar9 = local_70;
        plVar13 = (longlong *)0x0;
      }
      else {
        plVar13[4] = lVar8;
        piVar3 = DAT_1803b6d60;
        if (*DAT_1803b6d60 + 1 != 0) {
          *DAT_1803b6d60 = *DAT_1803b6d60 + 1;
        }
        plVar13[5] = (longlong)piVar3;
        plVar12 = (longlong *)PyUnicode_Join(DAT_1803b6c50,plVar13);
        if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
          (**(code **)(plVar13[1] + 0x30))(plVar13);
        }
        uVar6 = DAT_1803b6c50;
        if (plVar12 == (longlong *)0x0) {
          lVar8 = *(longlong *)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
            (**(code **)(plVar9[1] + 0x30))(plVar9);
          }
          uVar15 = 0x87;
          plVar9 = local_70;
          plVar13 = local_68;
        }
        else {
          lVar8 = FUN_18014c240(param_1);
          if (lVar8 == 0) {
            FUN_180373340(param_1,&local_60,DAT_1803b6d68);
            if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
              (**(code **)(plVar9[1] + 0x30))(plVar9);
            }
            if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
              (**(code **)(plVar12[1] + 0x30))(plVar12);
            }
            uVar15 = 0x87;
            lVar8 = local_60;
            plVar9 = local_70;
            plVar13 = local_68;
          }
          else {
            uVar15 = 0x87;
            *(undefined4 *)(plVar5 + 5) = 0x87;
            plVar13 = (longlong *)FUN_180376770(param_1,lVar8,local_70);
            if (plVar13 == (longlong *)0x0) {
              lVar8 = *(longlong *)(param_1 + 0x60);
              *(undefined8 *)(param_1 + 0x60) = 0;
              if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                (**(code **)(plVar9[1] + 0x30))(plVar9);
              }
              plVar9 = local_70;
              plVar13 = local_68;
              if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                (**(code **)(plVar12[1] + 0x30))(plVar12);
                plVar9 = local_70;
                plVar13 = local_68;
              }
            }
            else {
              *(undefined4 *)(plVar5 + 5) = 0x87;
              local_50 = uVar6;
              local_58 = plVar12;
              local_48 = plVar13;
              plVar14 = (longlong *)FUN_180378890(param_1,plVar9,&local_58);
              if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                (**(code **)(plVar9[1] + 0x30))(plVar9);
              }
              if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                (**(code **)(plVar12[1] + 0x30))(plVar12);
              }
              if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                (**(code **)(plVar13[1] + 0x30))(plVar13);
              }
              if (plVar14 != (longlong *)0x0) {
                plVar9 = (longlong *)FUN_18037f0c0();
                if ((-1 < (int)*plVar14) && (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                  (**(code **)(plVar14[1] + 0x30))(plVar14);
                }
                if (plVar9 != (longlong *)0x0) {
                  *(undefined4 *)(plVar5 + 5) = 0x87;
                  plVar13 = (longlong *)FUN_180376210(param_1,plVar9);
                  if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                    (**(code **)(plVar9[1] + 0x30))(plVar9);
                  }
                  if (plVar13 != (longlong *)0x0) {
                    iVar7 = FUN_180001510(plVar13);
                    piVar3 = DAT_1803b6d70;
                    if (iVar7 != -1) {
                      if (iVar7 == 1) {
                        iVar7 = (int)*plVar13 + 1;
                        plVar9 = plVar13;
                        if (iVar7 != 0) {
                          *(int *)plVar13 = iVar7;
                        }
LAB_1801527eb:
                        FUN_180001d80(param_1);
                        if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                          (**(code **)(plVar13[1] + 0x30))(plVar13);
                        }
                        if ((-1 < (int)*local_70) && (*local_70 = *local_70 + -1, *local_70 == 0)) {
                          (**(code **)(local_70[1] + 0x30))();
                        }
                        if ((int)*plVar1 < 0) {
                          return plVar9;
                        }
                        *plVar1 = *plVar1 + -1;
                        if (*plVar1 != 0) {
                          return plVar9;
                        }
                        (**(code **)(plVar1[1] + 0x30))(plVar1);
                        return plVar9;
                      }
                      lVar8 = *(longlong *)(param_1 + 0x10);
                      plVar12 = *(longlong **)(lVar8 + 0x41da8);
                      if (plVar12 == (longlong *)0x0) {
                        plVar12 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,3);
                      }
                      else {
                        *(longlong *)(lVar8 + 0x41da8) = plVar12[3];
                        *(int *)(lVar8 + 0x41e40) = *(int *)(lVar8 + 0x41e40) + -1;
                        *plVar12 = 1;
                      }
                      plVar9 = plVar12 + -2;
                      plVar12[3] = 0;
                      plVar12[4] = 0;
                      plVar12[5] = 0;
                      lVar8 = _PyThreadState_GetCurrent();
                      lVar8 = *(longlong *)(*(longlong *)(lVar8 + 0x10) + 0xd0);
                      puVar2 = *(undefined8 **)(lVar8 + 8);
                      *puVar2 = plVar9;
                      plVar12[-1] = plVar12[-1] & 3;
                      plVar12[-1] = plVar12[-1] | (ulonglong)puVar2;
                      *plVar9 = lVar8;
                      *(longlong **)(lVar8 + 8) = plVar9;
                      if (*piVar3 + 1 != 0) {
                        *piVar3 = *piVar3 + 1;
                      }
                      plVar12[3] = (longlong)piVar3;
                      lVar8 = FUN_180375520(param_1,plVar1,DAT_1803b6c50);
                      if (lVar8 == 0) {
                        lVar8 = *(longlong *)(param_1 + 0x60);
                        uVar15 = 0x88;
                        *(undefined8 *)(param_1 + 0x60) = 0;
                        plVar9 = local_70;
                        if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                          (**(code **)(plVar12[1] + 0x30))(plVar12);
                          plVar9 = local_70;
                        }
                        goto LAB_18015254b;
                      }
                      plVar12[4] = lVar8;
                      piVar3 = DAT_1803b6d78;
                      if (*DAT_1803b6d78 + 1 != 0) {
                        *DAT_1803b6d78 = *DAT_1803b6d78 + 1;
                      }
                      plVar12[5] = (longlong)piVar3;
                      plVar9 = (longlong *)PyUnicode_Join(DAT_1803b6c50,plVar12);
                      if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                        (**(code **)(plVar12[1] + 0x30))(plVar12);
                      }
                      if (plVar9 != (longlong *)0x0) goto LAB_1801527eb;
                    }
                    lVar8 = *(longlong *)(param_1 + 0x60);
                    uVar15 = 0x88;
                    *(undefined8 *)(param_1 + 0x60) = 0;
                    plVar9 = local_70;
                    goto LAB_18015254b;
                  }
                }
              }
              lVar8 = *(longlong *)(param_1 + 0x60);
              *(undefined8 *)(param_1 + 0x60) = 0;
              plVar9 = local_70;
              plVar13 = local_68;
            }
          }
        }
      }
    }
  }
LAB_18015254b:
  piVar3 = *(int **)(lVar8 + 0x28);
  if (piVar3 == (int *)0x0) {
    pcVar10 = (code *)FUN_18038e2a0(plVar5,uVar15);
    plVar12 = *(longlong **)(lVar8 + 0x28);
    pcVar11 = pcVar10;
    if (pcVar10 == _Py_NoneStruct_exref) {
      pcVar11 = (code *)0x0;
    }
    else {
LAB_1801529c5:
      pcVar10 = pcVar11;
      if ((pcVar11 != (code *)0x0) && (*(int *)pcVar11 + 1 != 0)) {
        *(int *)pcVar11 = *(int *)pcVar11 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar3 + 6) == plVar5) goto LAB_180152a0e;
    pcVar10 = (code *)FUN_18038e2a0(plVar5,uVar15);
    *(int **)(pcVar10 + 0x10) = piVar3;
    if (*piVar3 + 1 != 0) {
      *piVar3 = *piVar3 + 1;
    }
    plVar12 = *(longlong **)(lVar8 + 0x28);
    pcVar11 = pcVar10;
    if (pcVar10 != _Py_NoneStruct_exref) goto LAB_1801529c5;
    pcVar11 = (code *)0x0;
  }
  *(code **)(lVar8 + 0x28) = pcVar11;
  if (((plVar12 != (longlong *)0x0) && (-1 < (int)*plVar12)) &&
     (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
    (**(code **)(plVar12[1] + 0x30))();
  }
  if (((pcVar10 != (code *)0x0) && (-1 < *(int *)pcVar10)) &&
     (*(longlong *)pcVar10 = *(longlong *)pcVar10 + -1, *(longlong *)pcVar10 == 0)) {
    (**(code **)(*(longlong *)(pcVar10 + 8) + 0x30))(pcVar10);
  }
LAB_180152a0e:
  local_88 = plVar13;
  FUN_18036d0a0(plVar5,&DAT_1803948e8,plVar9,plVar1);
  plVar12 = DAT_1803b6a20;
  if (plVar5 == DAT_1803b6a20) {
    if ((-1 < (int)*DAT_1803b6a20) && (*DAT_1803b6a20 = *DAT_1803b6a20 + -1, *plVar12 == 0)) {
      (**(code **)(plVar12[1] + 0x30))();
    }
    DAT_1803b6a20 = (longlong *)0x0;
  }
  lVar4 = **(longlong **)(param_1 + 0x38);
  plVar5 = *(longlong **)(lVar4 + 0x30);
  **(longlong **)(param_1 + 0x38) = *(longlong *)(lVar4 + 8);
  *(undefined4 *)(plVar5 + 8) = 0xffffffff;
  plVar12 = (longlong *)plVar5[2];
  if (((plVar12 != (longlong *)0x0) && (plVar5[2] = 0, -1 < (int)*plVar12)) &&
     (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
    (**(code **)(plVar12[1] + 0x30))();
  }
  if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
    (**(code **)(plVar5[1] + 0x30))(plVar5);
  }
  plVar5[10] = 0;
  if (((plVar13 != (longlong *)0x0) && (-1 < (int)*plVar13)) &&
     (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
    (**(code **)(plVar13[1] + 0x30))(plVar13);
  }
  if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
    (**(code **)(plVar9[1] + 0x30))(plVar9);
  }
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))(plVar1);
  }
  plVar9 = *(longlong **)(param_1 + 0x60);
  *(longlong *)(param_1 + 0x60) = lVar8;
  if (((plVar9 != (longlong *)0x0) && (-1 < (int)*plVar9)) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)
     ) {
    (**(code **)(plVar9[1] + 0x30))();
  }
  return (longlong *)0x0;
}
