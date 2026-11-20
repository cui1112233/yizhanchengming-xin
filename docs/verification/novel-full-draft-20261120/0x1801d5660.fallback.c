/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x1801d5660 */

longlong * FUN_1801d5660(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  undefined8 *puVar2;
  int *piVar3;
  char cVar4;
  int iVar5;
  longlong *plVar6;
  longlong lVar7;
  longlong *plVar8;
  longlong *plVar9;
  longlong *plVar10;
  longlong *plVar11;
  longlong *plVar12;
  code *pcVar13;
  code *pcVar14;
  longlong *plVar15;
  undefined8 uVar16;
  longlong *plVar17;
  longlong *plVar18;

  plVar1 = (longlong *)*param_3;
  plVar11 = (longlong *)0x0;
  lVar7 = *(longlong *)(param_1 + 0x10);
  plVar18 = (longlong *)0x0;
  plVar17 = (longlong *)0x0;
  iVar5 = *(int *)(lVar7 + 0x42108);
  if (iVar5 == 0) {
    plVar6 = (longlong *)FUN_180001220(PyList_Type_exref);
  }
  else {
    *(int *)(lVar7 + 0x42108) = iVar5 + -1;
    plVar6 = *(longlong **)(lVar7 + 0x41e80 + (longlong)iVar5 * 8);
    *plVar6 = 1;
  }
  plVar9 = plVar6 + -2;
  plVar6[3] = 0;
  plVar6[2] = 0;
  plVar6[4] = 0;
  lVar7 = _PyThreadState_GetCurrent();
  lVar7 = *(longlong *)(*(longlong *)(lVar7 + 0x10) + 0xd0);
  puVar2 = *(undefined8 **)(lVar7 + 8);
  *puVar2 = plVar9;
  plVar6[-1] = plVar6[-1] & 3;
  plVar6[-1] = plVar6[-1] | (ulonglong)puVar2;
  plVar8 = DAT_1803bc918;
  *plVar9 = lVar7;
  *(longlong **)(lVar7 + 8) = plVar9;
  if (plVar8 == (longlong *)0x0) {
LAB_1801d573f:
    plVar8 = (longlong *)FUN_18036cd10(param_1,DAT_1803bc8e8,DAT_1803ca5e8,0x20);
    DAT_1803bc918 = plVar8;
  }
  else if ((1 < *plVar8) || (plVar8[2] != 0)) {
    if ((plVar8 != (longlong *)0x0) &&
       ((-1 < (int)*plVar8 && (*plVar8 = *plVar8 + -1, *plVar8 == 0)))) {
      (**(code **)(plVar8[1] + 0x30))();
    }
    goto LAB_1801d573f;
  }
  FUN_180001d00(param_1,plVar8);
  plVar9 = (longlong *)FUN_180375130(param_1,plVar1);
  if (plVar9 == (longlong *)0x0) {
    plVar15 = *(longlong **)(param_1 + 0x60);
    uVar16 = 0xb6;
    *(undefined8 *)(param_1 + 0x60) = 0;
    plVar17 = plVar11;
    plVar18 = plVar11;
  }
  else {
    plVar10 = (longlong *)FUN_18000d280();
    if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
      (**(code **)(plVar9[1] + 0x30))(plVar9);
    }
    if (plVar10 != (longlong *)0x0) {
      plVar9 = (longlong *)(**(code **)(plVar10[1] + 0xe0))(plVar10);
      plVar17 = plVar11;
      while (plVar18 = plVar9, plVar18 != (longlong *)0x0) {
        if (((plVar11 != (longlong *)0x0) && (-1 < (int)*plVar11)) &&
           (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
          (**(code **)(plVar11[1] + 0x30))(plVar11);
        }
        iVar5 = (int)*plVar18 + 1;
        if (iVar5 != 0) {
          *(int *)plVar18 = iVar5;
        }
        if (((plVar11 != (longlong *)0x0) && (-1 < (int)*plVar11)) &&
           (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
          (**(code **)(plVar11[1] + 0x30))();
        }
        plVar9 = (longlong *)FUN_180380b80(plVar18,DAT_1803ab258);
        if (plVar9 == (longlong *)0x0) {
          plVar15 = *(longlong **)(param_1 + 0x60);
          uVar16 = 0xb7;
          *(undefined8 *)(param_1 + 0x60) = 0;
          goto LAB_1801d5e01;
        }
        if (((plVar17 != (longlong *)0x0) && (-1 < (int)*plVar17)) &&
           (*plVar17 = *plVar17 + -1, *plVar17 == 0)) {
          (**(code **)(plVar17[1] + 0x30))();
        }
        piVar3 = DAT_1803bca98;
        lVar7 = *(longlong *)(param_1 + 0x10);
        plVar11 = *(longlong **)(lVar7 + 0x41da8);
        if (plVar11 == (longlong *)0x0) {
          plVar11 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,3);
        }
        else {
          *(longlong *)(lVar7 + 0x41da8) = plVar11[3];
          *(int *)(lVar7 + 0x41e40) = *(int *)(lVar7 + 0x41e40) + -1;
          *plVar11 = 1;
        }
        plVar17 = plVar11 + -2;
        plVar11[3] = 0;
        plVar11[4] = 0;
        plVar11[5] = 0;
        lVar7 = _PyThreadState_GetCurrent();
        lVar7 = *(longlong *)(*(longlong *)(lVar7 + 0x10) + 0xd0);
        puVar2 = *(undefined8 **)(lVar7 + 8);
        *puVar2 = plVar17;
        plVar11[-1] = plVar11[-1] & 3;
        plVar11[-1] = plVar11[-1] | (ulonglong)puVar2;
        *plVar17 = lVar7;
        *(longlong **)(lVar7 + 8) = plVar17;
        if (*piVar3 + 1 != 0) {
          *piVar3 = *piVar3 + 1;
        }
        plVar11[3] = (longlong)piVar3;
        lVar7 = FUN_180375520(param_1,plVar9,DAT_1803bc980);
        plVar17 = plVar9;
        if (lVar7 == 0) {
          plVar15 = *(longlong **)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          uVar16 = 0xb9;
          if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
            (**(code **)(plVar11[1] + 0x30))(plVar11);
          }
          goto LAB_1801d5e01;
        }
        plVar11[4] = lVar7;
        piVar3 = DAT_1803bcaa0;
        if (*DAT_1803bcaa0 + 1 != 0) {
          *DAT_1803bcaa0 = *DAT_1803bcaa0 + 1;
        }
        plVar11[5] = (longlong)piVar3;
        lVar7 = PyUnicode_Join(DAT_1803bc980,plVar11);
        if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
          (**(code **)(plVar11[1] + 0x30))(plVar11);
        }
        if (lVar7 == 0) {
          plVar15 = *(longlong **)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          uVar16 = 0xb9;
          goto LAB_1801d5e01;
        }
        plVar11 = (longlong *)FUN_180372160(param_1,4);
        *(longlong *)plVar11[3] = lVar7;
        piVar3 = DAT_1803bcaa8;
        if (*DAT_1803bcaa8 + 1 != 0) {
          *DAT_1803bcaa8 = *DAT_1803bcaa8 + 1;
        }
        *(int **)(plVar11[3] + 8) = piVar3;
        piVar3 = DAT_1803bca98;
        lVar7 = *(longlong *)(param_1 + 0x10);
        plVar12 = *(longlong **)(lVar7 + 0x41da8);
        if (plVar12 == (longlong *)0x0) {
          plVar12 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,3);
        }
        else {
          *(longlong *)(lVar7 + 0x41da8) = plVar12[3];
          *(int *)(lVar7 + 0x41e40) = *(int *)(lVar7 + 0x41e40) + -1;
          *plVar12 = 1;
        }
        plVar15 = plVar12 + -2;
        plVar12[3] = 0;
        plVar12[4] = 0;
        plVar12[5] = 0;
        lVar7 = _PyThreadState_GetCurrent();
        lVar7 = *(longlong *)(*(longlong *)(lVar7 + 0x10) + 0xd0);
        puVar2 = *(undefined8 **)(lVar7 + 8);
        *puVar2 = plVar15;
        plVar12[-1] = plVar12[-1] & 3;
        plVar12[-1] = plVar12[-1] | (ulonglong)puVar2;
        *plVar15 = lVar7;
        *(longlong **)(lVar7 + 8) = plVar15;
        if (*piVar3 + 1 != 0) {
          *piVar3 = *piVar3 + 1;
        }
        plVar12[3] = (longlong)piVar3;
        lVar7 = FUN_180375520(param_1,plVar9,DAT_1803bc980);
        if (lVar7 == 0) {
          uVar16 = 0xbb;
LAB_1801d5cf9:
          plVar15 = *(longlong **)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
            (**(code **)(plVar12[1] + 0x30))(plVar12);
          }
LAB_1801d5d1d:
          if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
            (**(code **)(plVar11[1] + 0x30))(plVar11);
          }
          goto LAB_1801d5e01;
        }
        plVar12[4] = lVar7;
        piVar3 = DAT_1803bcab0;
        if (*DAT_1803bcab0 + 1 != 0) {
          *DAT_1803bcab0 = *DAT_1803bcab0 + 1;
        }
        plVar12[5] = (longlong)piVar3;
        lVar7 = PyUnicode_Join(DAT_1803bc980,plVar12);
        if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
          (**(code **)(plVar12[1] + 0x30))(plVar12);
        }
        if (lVar7 == 0) {
          plVar15 = *(longlong **)(param_1 + 0x60);
          uVar16 = 0xbb;
          *(undefined8 *)(param_1 + 0x60) = 0;
          goto LAB_1801d5d1d;
        }
        *(longlong *)(plVar11[3] + 0x10) = lVar7;
        piVar3 = DAT_1803bcab8;
        lVar7 = *(longlong *)(param_1 + 0x10);
        plVar12 = *(longlong **)(lVar7 + 0x41da8);
        if (plVar12 == (longlong *)0x0) {
          plVar12 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,3);
        }
        else {
          *(longlong *)(lVar7 + 0x41da8) = plVar12[3];
          *(int *)(lVar7 + 0x41e40) = *(int *)(lVar7 + 0x41e40) + -1;
          *plVar12 = 1;
        }
        plVar15 = plVar12 + -2;
        plVar12[3] = 0;
        plVar12[4] = 0;
        plVar12[5] = 0;
        lVar7 = _PyThreadState_GetCurrent();
        lVar7 = *(longlong *)(*(longlong *)(lVar7 + 0x10) + 0xd0);
        puVar2 = *(undefined8 **)(lVar7 + 8);
        *puVar2 = plVar15;
        plVar12[-1] = plVar12[-1] & 3;
        plVar12[-1] = plVar12[-1] | (ulonglong)puVar2;
        *plVar15 = lVar7;
        *(longlong **)(lVar7 + 8) = plVar15;
        if (*piVar3 + 1 != 0) {
          *piVar3 = *piVar3 + 1;
        }
        plVar12[3] = (longlong)piVar3;
        lVar7 = FUN_180375520(param_1,plVar9,DAT_1803bc980);
        if (lVar7 == 0) {
          uVar16 = 0xbc;
          goto LAB_1801d5cf9;
        }
        plVar12[4] = lVar7;
        piVar3 = DAT_1803bcac0;
        if (*DAT_1803bcac0 + 1 != 0) {
          *DAT_1803bcac0 = *DAT_1803bcac0 + 1;
        }
        plVar12[5] = (longlong)piVar3;
        lVar7 = PyUnicode_Join(DAT_1803bc980,plVar12);
        if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
          (**(code **)(plVar12[1] + 0x30))(plVar12);
        }
        if (lVar7 == 0) {
          plVar15 = *(longlong **)(param_1 + 0x60);
          uVar16 = 0xbc;
          *(undefined8 *)(param_1 + 0x60) = 0;
          goto LAB_1801d5d1d;
        }
        *(longlong *)(plVar11[3] + 0x18) = lVar7;
        cVar4 = FUN_180372360(param_1,plVar6,plVar11);
        if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
          (**(code **)(plVar11[1] + 0x30))(plVar11);
        }
        if (cVar4 == '\0') {
          plVar15 = *(longlong **)(param_1 + 0x60);
          uVar16 = 0xb8;
          *(undefined8 *)(param_1 + 0x60) = 0;
          goto LAB_1801d5e01;
        }
        lVar7 = *(longlong *)(param_1 + 0x10);
        if ((((*(int *)(*(longlong *)(lVar7 + 0x60) + 0x4d4) != 0) || (*(int *)(lVar7 + 0x1a0) != 0)
             ) && (iVar5 = Py_MakePendingCalls(), iVar5 < 0)) &&
           (*(longlong *)(param_1 + 0x60) != 0)) {
          plVar15 = *(longlong **)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          goto LAB_1801d5dfc;
        }
        if (*(int *)(lVar7 + 0x174) != 0) {
          PyEval_SaveThread();
          PyEval_AcquireThread(param_1);
        }
        if (*(longlong *)(param_1 + 0x80) != 0) {
          plVar17 = *(longlong **)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x80) = 0;
          uVar16 = FUN_180373490(param_1);
          *(undefined8 *)(param_1 + 0x60) = uVar16;
          if (((plVar17 != (longlong *)0x0) && (-1 < (int)*plVar17)) &&
             (*plVar17 = *plVar17 + -1, *plVar17 == 0)) {
            (**(code **)(plVar17[1] + 0x30))(plVar17);
          }
          plVar15 = *(longlong **)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          goto LAB_1801d5dfc;
        }
        plVar9 = (longlong *)(**(code **)(plVar10[1] + 0xe0))(plVar10);
        plVar11 = plVar18;
      }
      plVar15 = *(longlong **)(param_1 + 0x60);
      if (plVar15 != (longlong *)0x0) {
        plVar18 = (longlong *)plVar15[1];
        if ((*(uint *)((longlong *)plVar15[1] + 0x15) & 0x40000000) == 0) {
          plVar18 = plVar15;
        }
        if ((plVar18 != *(longlong **)PyExc_StopIteration_exref) &&
           (((-1 < *(int *)(plVar18[1] + 0xa8) || ((*(uint *)(plVar18 + 0x15) & 0x40000000) == 0))
            || (cVar4 = FUN_180382640(), cVar4 != '\x01')))) {
          *(undefined8 *)(param_1 + 0x60) = 0;
          plVar9 = plVar17;
          plVar18 = plVar11;
LAB_1801d5dfc:
          uVar16 = 0xb6;
          plVar17 = plVar9;
LAB_1801d5e01:
          if (((plVar18 != (longlong *)0x0) && (-1 < (int)*plVar18)) &&
             (*plVar18 = *plVar18 + -1, *plVar18 == 0)) {
            (**(code **)(plVar18[1] + 0x30))(plVar18);
          }
          if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
            (**(code **)(plVar10[1] + 0x30))();
          }
          goto LAB_1801d5e45;
        }
        *(undefined8 *)(param_1 + 0x60) = 0;
        if ((-1 < (int)*plVar15) && (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
          (**(code **)(plVar15[1] + 0x30))(plVar15);
        }
      }
      if (((plVar11 != (longlong *)0x0) && (-1 < (int)*plVar11)) &&
         (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
        (**(code **)(plVar11[1] + 0x30))(plVar11);
      }
      if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
        (**(code **)(plVar10[1] + 0x30))(plVar10);
      }
      iVar5 = (int)*plVar6 + 1;
      if (iVar5 != 0) {
        *(int *)plVar6 = iVar5;
      }
      FUN_180001d80(param_1);
      if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
        (**(code **)(plVar6[1] + 0x30))(plVar6);
      }
      if (((plVar11 != (longlong *)0x0) && (-1 < (int)*plVar11)) &&
         (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
        (**(code **)(plVar11[1] + 0x30))(plVar11);
      }
      if (((plVar17 != (longlong *)0x0) && (-1 < (int)*plVar17)) &&
         (*plVar17 = *plVar17 + -1, *plVar17 == 0)) {
        (**(code **)(plVar17[1] + 0x30))(plVar17);
      }
      if ((int)*plVar1 < 0) {
        return plVar6;
      }
      *plVar1 = *plVar1 + -1;
      if (*plVar1 != 0) {
        return plVar6;
      }
      (**(code **)(plVar1[1] + 0x30))(plVar1);
      return plVar6;
    }
    plVar15 = *(longlong **)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
    uVar16 = 0xb6;
  }
LAB_1801d5e45:
  piVar3 = (int *)plVar15[5];
  if (piVar3 == (int *)0x0) {
    pcVar13 = (code *)FUN_18038e2a0(plVar8,uVar16);
    plVar11 = (longlong *)plVar15[5];
    if (pcVar13 == _Py_NoneStruct_exref) {
      pcVar14 = (code *)0x0;
    }
    else {
      pcVar14 = pcVar13;
      if ((pcVar13 != (code *)0x0) && (*(int *)pcVar13 + 1 != 0)) {
        *(int *)pcVar13 = *(int *)pcVar13 + 1;
      }
    }
    plVar15[5] = (longlong)pcVar14;
    if (((plVar11 != (longlong *)0x0) && (-1 < (int)*plVar11)) &&
       (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
      (**(code **)(plVar11[1] + 0x30))();
    }
    if (((pcVar13 != (code *)0x0) && (-1 < *(int *)pcVar13)) &&
       (*(longlong *)pcVar13 = *(longlong *)pcVar13 + -1, *(longlong *)pcVar13 == 0)) {
      (**(code **)(*(longlong *)(pcVar13 + 8) + 0x30))(pcVar13);
    }
  }
  else if (*(longlong **)(piVar3 + 6) != plVar8) {
    pcVar13 = (code *)FUN_18038e2a0(plVar8,uVar16);
    *(int **)(pcVar13 + 0x10) = piVar3;
    if (*piVar3 + 1 != 0) {
      *piVar3 = *piVar3 + 1;
    }
    plVar11 = (longlong *)plVar15[5];
    if (pcVar13 == _Py_NoneStruct_exref) {
      pcVar14 = (code *)0x0;
    }
    else {
      pcVar14 = pcVar13;
      if ((pcVar13 != (code *)0x0) && (*(int *)pcVar13 + 1 != 0)) {
        *(int *)pcVar13 = *(int *)pcVar13 + 1;
      }
    }
    plVar15[5] = (longlong)pcVar14;
    if (((plVar11 != (longlong *)0x0) && (-1 < (int)*plVar11)) &&
       (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
      (**(code **)(plVar11[1] + 0x30))();
    }
    if (((pcVar13 != (code *)0x0) && (-1 < *(int *)pcVar13)) &&
       (*(longlong *)pcVar13 = *(longlong *)pcVar13 + -1, *(longlong *)pcVar13 == 0)) {
      (**(code **)(*(longlong *)(pcVar13 + 8) + 0x30))(pcVar13);
    }
  }
  FUN_18036d0a0(plVar8,&DAT_180391c1c,plVar1,plVar6,plVar18,plVar17);
  plVar11 = DAT_1803bc918;
  if (plVar8 == DAT_1803bc918) {
    if ((-1 < (int)*DAT_1803bc918) && (*DAT_1803bc918 = *DAT_1803bc918 + -1, *plVar11 == 0)) {
      (**(code **)(plVar11[1] + 0x30))();
    }
    DAT_1803bc918 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
    (**(code **)(plVar6[1] + 0x30))(plVar6);
  }
  if (((plVar18 != (longlong *)0x0) && (-1 < (int)*plVar18)) &&
     (*plVar18 = *plVar18 + -1, *plVar18 == 0)) {
    (**(code **)(plVar18[1] + 0x30))(plVar18);
  }
  if (((plVar17 != (longlong *)0x0) && (-1 < (int)*plVar17)) &&
     (*plVar17 = *plVar17 + -1, *plVar17 == 0)) {
    (**(code **)(plVar17[1] + 0x30))(plVar17);
  }
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))(plVar1);
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong **)(param_1 + 0x60) = plVar15;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return (longlong *)0x0;
}
