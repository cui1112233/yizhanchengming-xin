/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x18014e500 */

int * FUN_18014e500(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  longlong lVar2;
  longlong *plVar3;
  char cVar4;
  int iVar5;
  longlong lVar6;
  undefined8 uVar7;
  longlong *plVar8;
  longlong *plVar9;
  longlong *plVar10;
  longlong *plVar11;
  int *piVar12;
  code *pcVar13;
  code *pcVar14;
  longlong *plVar15;
  longlong *plVar16;
  longlong *local_res18;
  longlong *local_res20;
  longlong *local_58;
  longlong *local_50;
  longlong *local_48;

  plVar3 = DAT_1803b6be8;
  plVar8 = (longlong *)*param_3;
  plVar1 = (longlong *)param_3[1];
  local_58 = (longlong *)0x0;
  local_res18 = DAT_180391d20;
  local_res20 = plVar8;
  if (DAT_1803b6be8 == (longlong *)0x0) {
LAB_18014e57b:
    DAT_1803b6be8 = (longlong *)FUN_18036cd10(param_1,DAT_1803b72d8,DAT_1803ca758,0x18);
  }
  else if ((1 < *DAT_1803b6be8) || (DAT_1803b6be8[2] != 0)) {
    if ((DAT_1803b6be8 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b6be8 && (*DAT_1803b6be8 = *DAT_1803b6be8 + -1, *plVar3 == 0)))) {
      (**(code **)(plVar3[1] + 0x30))(plVar3);
    }
    goto LAB_18014e57b;
  }
  plVar3 = DAT_1803b6be8;
  FUN_180001d00(param_1,DAT_1803b6be8);
  lVar6 = FUN_18014c130(param_1);
  if (lVar6 == 0) {
    FUN_180373340(param_1,&local_res18,DAT_1803b6cc0);
    uVar7 = 0x4e;
    plVar15 = local_res18;
    plVar8 = (longlong *)0x0;
  }
  else {
    *(undefined4 *)(plVar3 + 5) = 0x4e;
    local_50 = plVar8;
    local_48 = plVar1;
    plVar8 = (longlong *)FUN_180377750(param_1,lVar6,&local_50);
    if (plVar8 == (longlong *)0x0) {
      plVar15 = *(longlong **)(param_1 + 0x60);
      uVar7 = 0x4e;
      *(undefined8 *)(param_1 + 0x60) = 0;
      plVar8 = (longlong *)0x0;
    }
    else {
      plVar9 = (longlong *)FUN_18000d280();
      local_res18 = plVar9;
      if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
        (**(code **)(plVar8[1] + 0x30))(plVar8);
      }
      if (plVar9 == (longlong *)0x0) {
        plVar15 = *(longlong **)(param_1 + 0x60);
        uVar7 = 0x4e;
        *(undefined8 *)(param_1 + 0x60) = 0;
        plVar8 = (longlong *)0x0;
      }
      else {
        plVar8 = (longlong *)FUN_180372160(param_1,0);
        plVar16 = (longlong *)0x0;
        do {
          pcVar13 = *(code **)(plVar9[1] + 0xe0);
          if (pcVar13 == (code *)0x0) {
            PyErr_Format(*(undefined8 *)PyExc_TypeError_exref,"\'%s\' object is not an iterator",
                         *(undefined8 *)(plVar9[1] + 0x18));
LAB_18014e8e0:
            plVar15 = *(longlong **)(param_1 + 0x60);
            if (plVar15 != (longlong *)0x0) {
              plVar10 = (longlong *)plVar15[1];
              if ((*(uint *)((longlong *)plVar15[1] + 0x15) & 0x40000000) == 0) {
                plVar10 = plVar15;
              }
              if ((plVar10 != *(longlong **)PyExc_StopIteration_exref) &&
                 (((-1 < *(int *)(plVar10[1] + 0xa8) ||
                   ((*(uint *)(plVar10 + 0x15) & 0x40000000) == 0)) ||
                  (cVar4 = FUN_180382640(), cVar4 != '\x01')))) goto LAB_18014e938;
              *(undefined8 *)(param_1 + 0x60) = 0;
              if ((-1 < (int)*plVar15) && (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                (**(code **)(plVar15[1] + 0x30))(plVar15);
              }
            }
            iVar5 = (int)*plVar8 + 1;
            if (iVar5 != 0) {
              *(int *)plVar8 = iVar5;
            }
            if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
              (**(code **)(plVar9[1] + 0x30))(plVar9);
            }
            if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
              (**(code **)(plVar8[1] + 0x30))(plVar8);
            }
            if (((plVar16 != (longlong *)0x0) && (-1 < (int)*plVar16)) &&
               (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
              (**(code **)(plVar16[1] + 0x30))(plVar16);
            }
            if (((plVar16 != (longlong *)0x0) && (-1 < (int)*plVar16)) &&
               (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
              (**(code **)(plVar16[1] + 0x30))(plVar16);
            }
            iVar5 = FUN_180001510(plVar8);
            piVar12 = DAT_1803b6c50;
            if (iVar5 != -1) {
              if (iVar5 == 0) {
                if (*DAT_1803b6c50 + 1 != 0) {
                  *DAT_1803b6c50 = *DAT_1803b6c50 + 1;
                }
LAB_18014ec3c:
                FUN_180001d80(param_1);
                if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                  (**(code **)(plVar8[1] + 0x30))(plVar8);
                }
                if ((-1 < (int)*local_res20) &&
                   (*local_res20 = *local_res20 + -1, *local_res20 == 0)) {
                  (**(code **)(local_res20[1] + 0x30))();
                }
                if ((int)*plVar1 < 0) {
                  return piVar12;
                }
                *plVar1 = *plVar1 + -1;
                if (*plVar1 != 0) {
                  return piVar12;
                }
                (**(code **)(plVar1[1] + 0x30))();
                return piVar12;
              }
              piVar12 = (int *)FUN_18014b9b0(param_1,plVar8,DAT_1803ab260,0xffffffffffffffff);
              if (piVar12 != (int *)0x0) goto LAB_18014ec3c;
            }
            plVar15 = *(longlong **)(param_1 + 0x60);
            uVar7 = 0x4f;
            *(undefined8 *)(param_1 + 0x60) = 0;
            goto LAB_18014eab0;
          }
          plVar10 = (longlong *)(*pcVar13)(plVar9);
          plVar9 = local_res18;
          if (plVar10 == (longlong *)0x0) goto LAB_18014e8e0;
          if (((plVar16 != (longlong *)0x0) && (-1 < (int)*plVar16)) &&
             (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
            (**(code **)(plVar16[1] + 0x30))();
          }
          iVar5 = (int)*plVar10 + 1;
          if (iVar5 != 0) {
            *(int *)plVar10 = iVar5;
          }
          if (((plVar16 != (longlong *)0x0) && (-1 < (int)*plVar16)) &&
             (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
            (**(code **)(plVar16[1] + 0x30))();
          }
          plVar9 = (longlong *)FUN_18037f0c0();
          plVar16 = plVar10;
          if (plVar9 == (longlong *)0x0) goto LAB_18014e881;
          *(undefined4 *)(plVar3 + 5) = 0x4e;
          plVar11 = (longlong *)FUN_180376210(param_1,plVar9);
          if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
            (**(code **)(plVar9[1] + 0x30))(plVar9);
          }
          if (plVar11 == (longlong *)0x0) goto LAB_18014e881;
          iVar5 = FUN_180001510(plVar11);
          if (iVar5 == -1) {
            plVar15 = *(longlong **)(param_1 + 0x60);
            *(undefined8 *)(param_1 + 0x60) = 0;
            if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
              (**(code **)(plVar11[1] + 0x30))(plVar11);
            }
            goto LAB_18014e940;
          }
          if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
            (**(code **)(plVar11[1] + 0x30))(plVar11);
          }
          if (iVar5 != 0) {
            plVar9 = (longlong *)FUN_18037f0c0();
            if (plVar9 == (longlong *)0x0) goto LAB_18014e881;
            *(undefined4 *)(plVar3 + 5) = 0x4e;
            lVar6 = FUN_180376210(param_1,plVar9);
            if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
              (**(code **)(plVar9[1] + 0x30))(plVar9);
            }
            if (lVar6 == 0) goto LAB_18014e881;
            lVar2 = plVar8[2];
            cVar4 = FUN_1803722b0(plVar8,lVar2 + 1);
            if (cVar4 == '\0') goto LAB_18014e881;
            *(longlong *)(plVar8[3] + lVar2 * 8) = lVar6;
          }
          lVar6 = *(longlong *)(param_1 + 0x10);
          if ((((*(int *)(*(longlong *)(lVar6 + 0x60) + 0x4d4) != 0) ||
               (*(int *)(lVar6 + 0x1a0) != 0)) && (iVar5 = Py_MakePendingCalls(), iVar5 < 0)) &&
             (*(longlong *)(param_1 + 0x60) != 0)) goto LAB_18014e881;
          if (*(int *)(lVar6 + 0x174) != 0) {
            PyEval_SaveThread();
            PyEval_AcquireThread(param_1);
          }
          plVar9 = local_res18;
        } while (*(longlong *)(param_1 + 0x80) == 0);
        plVar9 = *(longlong **)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x80) = 0;
        uVar7 = FUN_180373490(param_1);
        *(undefined8 *)(param_1 + 0x60) = uVar7;
        if (((plVar9 != (longlong *)0x0) && (-1 < (int)*plVar9)) &&
           (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
          (**(code **)(plVar9[1] + 0x30))(plVar9);
        }
LAB_18014e881:
        plVar15 = *(longlong **)(param_1 + 0x60);
LAB_18014e938:
        *(undefined8 *)(param_1 + 0x60) = 0;
        plVar10 = plVar16;
LAB_18014e940:
        if ((-1 < (int)*local_res18) && (*local_res18 = *local_res18 + -1, *local_res18 == 0)) {
          (**(code **)(local_res18[1] + 0x30))();
        }
        if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
          (**(code **)(plVar8[1] + 0x30))(plVar8);
        }
        if (((plVar10 != (longlong *)0x0) && (-1 < (int)*plVar10)) &&
           (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
          (**(code **)(plVar10[1] + 0x30))(plVar10);
        }
        if (((plVar10 != (longlong *)0x0) && (-1 < (int)*plVar10)) &&
           (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
          (**(code **)(plVar10[1] + 0x30))(plVar10);
        }
        uVar7 = 0x4e;
        plVar8 = local_58;
      }
    }
  }
LAB_18014eab0:
  piVar12 = (int *)plVar15[5];
  if (piVar12 == (int *)0x0) {
    pcVar13 = (code *)FUN_18038e2a0(plVar3,uVar7);
    plVar9 = (longlong *)plVar15[5];
    pcVar14 = pcVar13;
    if (pcVar13 == _Py_NoneStruct_exref) {
      pcVar14 = (code *)0x0;
    }
    else {
LAB_18014eb10:
      pcVar13 = pcVar14;
      if ((pcVar14 != (code *)0x0) && (*(int *)pcVar14 + 1 != 0)) {
        *(int *)pcVar14 = *(int *)pcVar14 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar12 + 6) == plVar3) goto LAB_18014eb59;
    pcVar13 = (code *)FUN_18038e2a0(plVar3,uVar7);
    *(int **)(pcVar13 + 0x10) = piVar12;
    if (*piVar12 + 1 != 0) {
      *piVar12 = *piVar12 + 1;
    }
    plVar9 = (longlong *)plVar15[5];
    pcVar14 = pcVar13;
    if (pcVar13 != _Py_NoneStruct_exref) goto LAB_18014eb10;
    pcVar14 = (code *)0x0;
  }
  plVar15[5] = (longlong)pcVar14;
  if (((plVar9 != (longlong *)0x0) && (-1 < (int)*plVar9)) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)
     ) {
    (**(code **)(plVar9[1] + 0x30))();
  }
  if (((pcVar13 != (code *)0x0) && (-1 < *(int *)pcVar13)) &&
     (*(longlong *)pcVar13 = *(longlong *)pcVar13 + -1, *(longlong *)pcVar13 == 0)) {
    (**(code **)(*(longlong *)(pcVar13 + 8) + 0x30))(pcVar13);
  }
LAB_18014eb59:
  plVar16 = local_res20;
  FUN_18036d0a0(plVar3,&DAT_1803948e8,local_res20,plVar1,plVar8);
  plVar9 = DAT_1803b6be8;
  if (plVar3 == DAT_1803b6be8) {
    if ((-1 < (int)*DAT_1803b6be8) && (*DAT_1803b6be8 = *DAT_1803b6be8 + -1, *plVar9 == 0)) {
      (**(code **)(plVar9[1] + 0x30))();
    }
    DAT_1803b6be8 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if (((plVar8 != (longlong *)0x0) && (-1 < (int)*plVar8)) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)
     ) {
    (**(code **)(plVar8[1] + 0x30))(plVar8);
  }
  if ((-1 < (int)*plVar16) && (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
    (**(code **)(plVar16[1] + 0x30))(plVar16);
  }
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))(plVar1);
  }
  plVar8 = *(longlong **)(param_1 + 0x60);
  *(longlong **)(param_1 + 0x60) = plVar15;
  if (((plVar8 != (longlong *)0x0) && (-1 < (int)*plVar8)) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)
     ) {
    (**(code **)(plVar8[1] + 0x30))();
  }
  return (int *)0x0;
}
