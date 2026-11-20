/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x1800d5c20 */

longlong * FUN_1800d5c20(longlong param_1,longlong param_2,undefined8 *param_3)

{
  longlong *plVar1;
  undefined8 *puVar2;
  code *pcVar3;
  int *piVar4;
  longlong lVar5;
  longlong *plVar6;
  int iVar7;
  uint uVar8;
  longlong *plVar9;
  longlong *plVar10;
  longlong lVar11;
  longlong *plVar12;
  longlong *plVar13;
  code *pcVar14;
  code *pcVar15;
  char *pcVar16;
  undefined8 uVar17;
  undefined8 uVar18;
  longlong *plVar19;
  char cVar20;
  longlong *local_res18;
  longlong local_res20;

  plVar6 = DAT_1803b29f8;
  plVar1 = (longlong *)*param_3;
  plVar19 = (longlong *)param_3[1];
  local_res20 = DAT_180391d00;
  if (DAT_1803b29f8 == (longlong *)0x0) {
LAB_1800d5c8f:
    DAT_1803b29f8 = (longlong *)FUN_18036cd10(param_1,DAT_1803b28f8,DAT_1803ca7a0,0x18);
  }
  else if ((1 < *DAT_1803b29f8) || (DAT_1803b29f8[2] != 0)) {
    if ((DAT_1803b29f8 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b29f8 && (*DAT_1803b29f8 = *DAT_1803b29f8 + -1, *plVar6 == 0)))) {
      (**(code **)(plVar6[1] + 0x30))(plVar6);
    }
    goto LAB_1800d5c8f;
  }
  plVar6 = DAT_1803b29f8;
  FUN_180001d00(param_1,DAT_1803b29f8);
  iVar7 = FUN_180001510(plVar19);
  pcVar15 = (code *)0x0;
  if (iVar7 == -1) {
    lVar11 = *(longlong *)(param_1 + 0x60);
    uVar17 = 0x1e1;
    *(undefined8 *)(param_1 + 0x60) = 0;
  }
  else {
    uVar8 = (uint)(iVar7 != 0);
    if (iVar7 == 0) {
LAB_1800d5e01:
      if (uVar8 == 1) {
        plVar9 = (longlong *)FUN_180372690(param_1,plVar19);
        if (plVar9 == (longlong *)0x0) {
          lVar11 = *(longlong *)(param_1 + 0x60);
          uVar17 = 0x1e2;
          *(undefined8 *)(param_1 + 0x60) = 0;
        }
        else {
          if ((-1 < (int)*plVar19) && (*plVar19 = *plVar19 + -1, *plVar19 == 0)) {
            (**(code **)(plVar19[1] + 0x30))(plVar19);
          }
          local_res18 = (longlong *)FUN_1800c6da0(param_1,plVar9,DAT_1803ab250);
          plVar19 = plVar9;
          if (local_res18 == (longlong *)0x0) {
            lVar11 = *(longlong *)(param_1 + 0x60);
            uVar17 = 0x1e4;
            *(undefined8 *)(param_1 + 0x60) = 0;
          }
          else {
            lVar11 = *(longlong *)(param_1 + 0x10);
            plVar10 = *(longlong **)(lVar11 + 0x41da0);
            if (plVar10 == (longlong *)0x0) {
              plVar10 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,2);
            }
            else {
              *(longlong *)(lVar11 + 0x41da0) = plVar10[3];
              *(int *)(lVar11 + 0x41e3c) = *(int *)(lVar11 + 0x41e3c) + -1;
              *plVar10 = 1;
            }
            plVar10[3] = 0;
            plVar12 = plVar10 + -2;
            plVar10[4] = 0;
            lVar11 = _PyThreadState_GetCurrent();
            lVar11 = *(longlong *)(*(longlong *)(lVar11 + 0x10) + 0xd0);
            puVar2 = *(undefined8 **)(lVar11 + 8);
            *puVar2 = plVar12;
            plVar10[-1] = plVar10[-1] & 3;
            plVar10[-1] = plVar10[-1] | (ulonglong)puVar2;
            *plVar12 = lVar11;
            *(longlong **)(lVar11 + 8) = plVar12;
            plVar10[3] = (longlong)local_res18;
            uVar17 = DAT_1803b2fd0;
            plVar12 = (longlong *)FUN_1800c6da0(param_1,plVar9,DAT_1803ab250);
            if (plVar12 == (longlong *)0x0) {
LAB_1800d600c:
              lVar11 = *(longlong *)(param_1 + 0x60);
              *(undefined8 *)(param_1 + 0x60) = 0;
            }
            else {
              plVar13 = (longlong *)FUN_180024e50(param_1,plVar12,DAT_1803b2fd0);
              if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                (**(code **)(plVar12[1] + 0x30))(plVar12);
              }
              if (plVar13 == (longlong *)0x0) goto LAB_1800d600c;
              plVar12 = (longlong *)FUN_180380f00(plVar13,DAT_1803b3230);
              if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                (**(code **)(plVar13[1] + 0x30))(plVar13);
              }
              if (plVar12 == (longlong *)0x0) goto LAB_1800d600c;
              if (*(longlong *)(*(longlong *)(param_2 + 0xe0) + 0x10) != 0) {
                plVar13 = (longlong *)FUN_180381110(plVar12);
                if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                  (**(code **)(plVar12[1] + 0x30))(plVar12);
                }
                if (plVar13 == (longlong *)0x0) goto LAB_1800d600c;
                lVar11 = _PyDict_NewPresized(1);
                PyDict_SetItem(lVar11,uVar17,plVar13);
                if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                  (**(code **)(plVar13[1] + 0x30))(plVar13);
                }
                plVar10[4] = lVar11;
                local_res18 = plVar10;
                plVar10 = (longlong *)FUN_180096610(param_1,&local_res18);
                if (plVar10 == (longlong *)0x0) {
                  lVar11 = *(longlong *)(param_1 + 0x60);
                  uVar17 = 0x1e3;
                  *(undefined8 *)(param_1 + 0x60) = 0;
                  goto LAB_1800d6232;
                }
                pcVar14 = (code *)plVar9[1];
                if ((*(longlong *)(pcVar14 + 0x70) == 0) ||
                   (pcVar3 = *(code **)(*(longlong *)(pcVar14 + 0x70) + 0x10), pcVar3 == (code *)0x0
                   )) {
                  if (*(longlong *)(pcVar14 + 0x68) == 0) {
                    uVar17 = *(undefined8 *)PyExc_TypeError_exref;
LAB_1800d6190:
                    uVar18 = *(undefined8 *)(pcVar14 + 0x18);
                    pcVar16 = "\'%s\' object does not support item assignment";
LAB_1800d619b:
                    PyErr_Format(uVar17,pcVar16,uVar18);
                  }
                  else {
                    lVar11 = *(longlong *)(*(longlong *)(DAT_1803ab250 + 8) + 0x60);
                    if ((lVar11 == 0) || (*(longlong *)(lVar11 + 0x108) == 0)) {
                      uVar17 = *(undefined8 *)PyExc_TypeError_exref;
                      if (*(longlong *)(*(longlong *)(pcVar14 + 0x68) + 0x28) == 0)
                      goto LAB_1800d6190;
                      uVar18 = *(undefined8 *)(*(longlong *)(DAT_1803ab250 + 8) + 0x18);
                      pcVar16 = "sequence index must be integer, not \'%s\'";
                      goto LAB_1800d619b;
                    }
                    lVar11 = PyNumber_AsSsize_t(DAT_1803ab250,*(undefined8 *)PyExc_IndexError_exref)
                    ;
                    if ((lVar11 != -1) || (*(longlong *)(param_1 + 0x60) == 0)) {
                      cVar20 = FUN_1803737c0(plVar9,lVar11,plVar10);
                      goto LAB_1800d61a4;
                    }
                  }
                  cVar20 = '\0';
                }
                else if (pcVar14 == PyList_Type_exref) {
                  plVar12 = *(longlong **)plVar9[3];
                  iVar7 = (int)*plVar10 + 1;
                  if (iVar7 != 0) {
                    *(int *)plVar10 = iVar7;
                  }
                  *(longlong **)plVar9[3] = plVar10;
                  if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                    (**(code **)(plVar12[1] + 0x30))();
                  }
                  cVar20 = '\x01';
                }
                else {
                  iVar7 = (*pcVar3)(plVar9,DAT_1803ab250,plVar10);
                  cVar20 = iVar7 != -1;
                }
LAB_1800d61a4:
                if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                  (**(code **)(plVar10[1] + 0x30))(plVar10);
                }
                if (cVar20 != '\0') {
LAB_1800d63e1:
                  iVar7 = (int)*plVar19 + 1;
                  if (iVar7 != 0) {
                    *(int *)plVar19 = iVar7;
                  }
                  FUN_180001d80(param_1);
                  if ((-1 < (int)*plVar19) && (*plVar19 = *plVar19 + -1, *plVar19 == 0)) {
                    (**(code **)(plVar19[1] + 0x30))(plVar19);
                  }
                  if ((int)*plVar1 < 0) {
                    return plVar19;
                  }
                  *plVar1 = *plVar1 + -1;
                  if (*plVar1 != 0) {
                    return plVar19;
                  }
                  (**(code **)(plVar1[1] + 0x30))(plVar1);
                  return plVar19;
                }
                lVar11 = *(longlong *)(param_1 + 0x60);
                uVar17 = 0x1e3;
                *(undefined8 *)(param_1 + 0x60) = 0;
                goto LAB_1800d6232;
              }
              if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                (**(code **)(plVar12[1] + 0x30))(plVar12);
              }
              FUN_180371050(param_1,&local_res20,DAT_1803b3238);
              lVar11 = local_res20;
              FUN_18000d1d0(param_1,local_res20);
            }
            uVar17 = 0x1e5;
            if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
              (**(code **)(plVar10[1] + 0x30))(plVar10);
            }
          }
        }
      }
      else {
        if (plVar19 != (longlong *)0x0) goto LAB_1800d63e1;
        plVar9 = (longlong *)
                 PyUnicode_FromFormat
                           ("cannot access local variable \'%U\' where it is not associated with a value"
                            ,DAT_1803b2c08);
        lVar11 = FUN_180376770(param_1,*(undefined8 *)PyExc_UnboundLocalError_exref,plVar9);
        if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
          (**(code **)(plVar9[1] + 0x30))(plVar9);
        }
        FUN_18000d1d0(param_1,lVar11);
        uVar17 = 0x1e7;
      }
    }
    else {
      plVar9 = (longlong *)FUN_1800c6da0(param_1,plVar19,DAT_1803ab250);
      if (plVar9 == (longlong *)0x0) {
        lVar11 = *(longlong *)(param_1 + 0x60);
        uVar17 = 0x1e1;
        *(undefined8 *)(param_1 + 0x60) = 0;
      }
      else {
        plVar10 = (longlong *)FUN_18037f0c0();
        if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
          (**(code **)(plVar9[1] + 0x30))(plVar9);
        }
        uVar17 = 0x1e1;
        if (plVar10 == (longlong *)0x0) {
          lVar11 = *(longlong *)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
        }
        else {
          *(undefined4 *)(plVar6 + 5) = 0x1e1;
          plVar9 = (longlong *)FUN_1803770d0(param_1,plVar10,DAT_1803b3228);
          if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
            (**(code **)(plVar10[1] + 0x30))(plVar10);
          }
          if (plVar9 == (longlong *)0x0) {
            lVar11 = *(longlong *)(param_1 + 0x60);
            *(undefined8 *)(param_1 + 0x60) = 0;
          }
          else {
            uVar8 = FUN_1803836d0(plVar9,DAT_1803b2fc8);
            local_res18 = (longlong *)CONCAT44(local_res18._4_4_,uVar8);
            if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
              (**(code **)(plVar9[1] + 0x30))(plVar9);
              uVar8 = (uint)local_res18;
            }
            if (uVar8 != 0xffffffff) goto LAB_1800d5e01;
            lVar11 = *(longlong *)(param_1 + 0x60);
            *(undefined8 *)(param_1 + 0x60) = 0;
          }
        }
      }
    }
  }
LAB_1800d6232:
  piVar4 = *(int **)(lVar11 + 0x28);
  if (piVar4 == (int *)0x0) {
    pcVar14 = (code *)FUN_18038e2a0(plVar6,uVar17);
    plVar9 = *(longlong **)(lVar11 + 0x28);
    if (pcVar14 != _Py_NoneStruct_exref) {
LAB_1800d628e:
      pcVar15 = pcVar14;
      pcVar14 = pcVar15;
      if ((pcVar15 != (code *)0x0) && (*(int *)pcVar15 + 1 != 0)) {
        *(int *)pcVar15 = *(int *)pcVar15 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar4 + 6) == plVar6) goto LAB_1800d62d7;
    pcVar14 = (code *)FUN_18038e2a0(plVar6,uVar17);
    *(int **)(pcVar14 + 0x10) = piVar4;
    if (*piVar4 + 1 != 0) {
      *piVar4 = *piVar4 + 1;
    }
    plVar9 = *(longlong **)(lVar11 + 0x28);
    if (pcVar14 != _Py_NoneStruct_exref) goto LAB_1800d628e;
  }
  *(code **)(lVar11 + 0x28) = pcVar15;
  if (((plVar9 != (longlong *)0x0) && (-1 < (int)*plVar9)) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)
     ) {
    (**(code **)(plVar9[1] + 0x30))();
  }
  if (((pcVar14 != (code *)0x0) && (-1 < *(int *)pcVar14)) &&
     (*(longlong *)pcVar14 = *(longlong *)pcVar14 + -1, *(longlong *)pcVar14 == 0)) {
    (**(code **)(*(longlong *)(pcVar14 + 8) + 0x30))(pcVar14);
  }
LAB_1800d62d7:
  FUN_18036d0a0(plVar6,&DAT_180394af8,plVar1,plVar19,*(undefined8 *)(param_2 + 0xe0));
  plVar9 = DAT_1803b29f8;
  if (plVar6 == DAT_1803b29f8) {
    if ((-1 < (int)*DAT_1803b29f8) && (*DAT_1803b29f8 = *DAT_1803b29f8 + -1, *plVar9 == 0)) {
      (**(code **)(plVar9[1] + 0x30))();
    }
    DAT_1803b29f8 = (longlong *)0x0;
  }
  lVar5 = **(longlong **)(param_1 + 0x38);
  plVar6 = *(longlong **)(lVar5 + 0x30);
  **(longlong **)(param_1 + 0x38) = *(longlong *)(lVar5 + 8);
  *(undefined4 *)(plVar6 + 8) = 0xffffffff;
  plVar9 = (longlong *)plVar6[2];
  if (((plVar9 != (longlong *)0x0) && (plVar6[2] = 0, -1 < (int)*plVar9)) &&
     (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
    (**(code **)(plVar9[1] + 0x30))();
  }
  if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
    (**(code **)(plVar6[1] + 0x30))(plVar6);
  }
  plVar6[10] = 0;
  if (((plVar19 != (longlong *)0x0) && (-1 < (int)*plVar19)) &&
     (*plVar19 = *plVar19 + -1, *plVar19 == 0)) {
    (**(code **)(plVar19[1] + 0x30))(plVar19);
  }
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))(plVar1);
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong *)(param_1 + 0x60) = lVar11;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return (longlong *)0x0;
}
