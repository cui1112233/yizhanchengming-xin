/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x1801d6160 */

longlong FUN_1801d6160(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  longlong lVar2;
  undefined8 *puVar3;
  int *piVar4;
  longlong *plVar5;
  char cVar6;
  longlong *plVar7;
  longlong lVar8;
  longlong *plVar9;
  longlong *plVar10;
  code *pcVar11;
  code *pcVar12;
  undefined8 uVar13;
  longlong *plVar14;
  longlong local_res8;

  plVar5 = DAT_1803bc888;
  plVar14 = (longlong *)0x0;
  plVar1 = (longlong *)*param_3;
  local_res8 = DAT_180391dd8;
  if (DAT_1803bc888 == (longlong *)0x0) {
LAB_1801d61c1:
    DAT_1803bc888 = (longlong *)FUN_18036cd10(param_1,DAT_1803bc8b0,DAT_1803ca5e8,0x10);
  }
  else if ((1 < *DAT_1803bc888) || (DAT_1803bc888[2] != 0)) {
    if ((DAT_1803bc888 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803bc888 && (*DAT_1803bc888 = *DAT_1803bc888 + -1, *plVar5 == 0)))) {
      (**(code **)(plVar5[1] + 0x30))(plVar5);
    }
    goto LAB_1801d61c1;
  }
  plVar5 = DAT_1803bc888;
  FUN_180001d00(param_1,DAT_1803bc888);
  piVar4 = DAT_1803bcac8;
  plVar7 = (longlong *)FUN_180372160(param_1,0x11);
  if (*piVar4 + 1 != 0) {
    *piVar4 = *piVar4 + 1;
  }
  *(int **)plVar7[3] = piVar4;
  piVar4 = DAT_1803bcad0;
  lVar8 = *(longlong *)(param_1 + 0x10);
  plVar10 = *(longlong **)(lVar8 + 0x41db8);
  if (plVar10 == (longlong *)0x0) {
    plVar10 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,5);
  }
  else {
    *(longlong *)(lVar8 + 0x41db8) = plVar10[3];
    *(int *)(lVar8 + 0x41e48) = *(int *)(lVar8 + 0x41e48) + -1;
    *plVar10 = 1;
  }
  plVar9 = plVar10 + -2;
  plVar10[3] = 0;
  plVar10[4] = 0;
  plVar10[5] = 0;
  plVar10[6] = 0;
  plVar10[7] = 0;
  lVar8 = _PyThreadState_GetCurrent();
  lVar8 = *(longlong *)(lVar8 + 0x10);
  lVar2 = *(longlong *)(lVar8 + 0xd0);
  puVar3 = *(undefined8 **)(lVar2 + 8);
  *puVar3 = plVar9;
  plVar10[-1] = plVar10[-1] & 3;
  plVar10[-1] = plVar10[-1] | (ulonglong)puVar3;
  *plVar9 = lVar2;
  *(longlong **)(lVar2 + 8) = plVar9;
  if (*piVar4 + 1 != 0) {
    *piVar4 = *piVar4 + 1;
  }
  plVar10[3] = (longlong)piVar4;
  plVar9 = (longlong *)FUN_18037f0c0(lVar8,plVar1,DAT_1803bc998);
  if (plVar9 == (longlong *)0x0) {
LAB_1801d635b:
    lVar8 = *(longlong *)(param_1 + 0x60);
    uVar13 = 0xc4;
    *(undefined8 *)(param_1 + 0x60) = 0;
    if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
      (**(code **)(plVar10[1] + 0x30))(plVar10);
    }
LAB_1801d6515:
    if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
      (**(code **)(plVar7[1] + 0x30))(plVar7);
    }
  }
  else {
    lVar8 = FUN_180375520(param_1,plVar9,DAT_1803bc980);
    if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
      (**(code **)(plVar9[1] + 0x30))(plVar9);
    }
    if (lVar8 == 0) goto LAB_1801d635b;
    plVar10[4] = lVar8;
    piVar4 = DAT_1803bc9a0;
    if (*DAT_1803bc9a0 + 1 != 0) {
      *DAT_1803bc9a0 = *DAT_1803bc9a0 + 1;
    }
    plVar10[5] = (longlong)piVar4;
    plVar9 = (longlong *)FUN_18037f0c0(piVar4,plVar1,DAT_1803bc9a8);
    if (plVar9 == (longlong *)0x0) goto LAB_1801d635b;
    lVar8 = FUN_180375520(param_1,plVar9,DAT_1803bc980);
    if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
      (**(code **)(plVar9[1] + 0x30))(plVar9);
    }
    if (lVar8 == 0) goto LAB_1801d635b;
    plVar10[6] = lVar8;
    piVar4 = DAT_1803bc9b0;
    if (*DAT_1803bc9b0 + 1 != 0) {
      *DAT_1803bc9b0 = *DAT_1803bc9b0 + 1;
    }
    plVar10[7] = (longlong)piVar4;
    lVar8 = PyUnicode_Join(DAT_1803bc980,plVar10);
    if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
      (**(code **)(plVar10[1] + 0x30))(plVar10);
    }
    if (lVar8 == 0) {
      uVar13 = 0xc4;
      goto LAB_1801d650d;
    }
    *(longlong *)(plVar7[3] + 8) = lVar8;
    piVar4 = DAT_1803bcad8;
    if (*DAT_1803bcad8 + 1 != 0) {
      *DAT_1803bcad8 = *DAT_1803bcad8 + 1;
    }
    *(int **)(plVar7[3] + 0x10) = piVar4;
    piVar4 = DAT_1803bc980;
    if (*DAT_1803bc980 + 1 != 0) {
      *DAT_1803bc980 = *DAT_1803bc980 + 1;
    }
    *(int **)(plVar7[3] + 0x18) = piVar4;
    lVar8 = FUN_1801d3800(param_1);
    if (lVar8 == 0) {
      FUN_180373340(param_1,&local_res8,DAT_1803bcae0);
      uVar13 = 199;
      lVar8 = local_res8;
      goto LAB_1801d6515;
    }
    uVar13 = 199;
    *(undefined4 *)(plVar5 + 5) = 199;
    lVar8 = FUN_180376210(param_1,lVar8);
    if (lVar8 == 0) {
LAB_1801d650d:
      lVar8 = *(longlong *)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      goto LAB_1801d6515;
    }
    *(longlong *)(plVar7[3] + 0x20) = lVar8;
    piVar4 = DAT_1803bc980;
    if (*DAT_1803bc980 + 1 != 0) {
      *DAT_1803bc980 = *DAT_1803bc980 + 1;
    }
    *(int **)(plVar7[3] + 0x28) = piVar4;
    piVar4 = DAT_1803bc970;
    if (*DAT_1803bc970 + 1 != 0) {
      *DAT_1803bc970 = *DAT_1803bc970 + 1;
    }
    *(int **)(plVar7[3] + 0x30) = piVar4;
    lVar8 = FUN_18037f0c0(piVar4,plVar1,DAT_1803bc978);
    if (lVar8 == 0) {
      uVar13 = 0xca;
      goto LAB_1801d650d;
    }
    *(longlong *)(plVar7[3] + 0x38) = lVar8;
    piVar4 = DAT_1803bc980;
    if (*DAT_1803bc980 + 1 != 0) {
      *DAT_1803bc980 = *DAT_1803bc980 + 1;
    }
    *(int **)(plVar7[3] + 0x40) = piVar4;
    piVar4 = DAT_1803bcae8;
    if (*DAT_1803bcae8 + 1 != 0) {
      *DAT_1803bcae8 = *DAT_1803bcae8 + 1;
    }
    *(int **)(plVar7[3] + 0x48) = piVar4;
    lVar8 = FUN_18037f0c0(piVar4,plVar1,DAT_1803bcaf0);
    if (lVar8 == 0) {
      uVar13 = 0xcd;
      goto LAB_1801d650d;
    }
    *(longlong *)(plVar7[3] + 0x50) = lVar8;
    piVar4 = DAT_1803bc980;
    if (*DAT_1803bc980 + 1 != 0) {
      *DAT_1803bc980 = *DAT_1803bc980 + 1;
    }
    *(int **)(plVar7[3] + 0x58) = piVar4;
    piVar4 = DAT_1803bca68;
    if (*DAT_1803bca68 + 1 != 0) {
      *DAT_1803bca68 = *DAT_1803bca68 + 1;
    }
    *(int **)(plVar7[3] + 0x60) = piVar4;
    piVar4 = DAT_1803bcaf8;
    if (*DAT_1803bcaf8 + 1 != 0) {
      *DAT_1803bcaf8 = *DAT_1803bcaf8 + 1;
    }
    *(int **)(plVar7[3] + 0x68) = piVar4;
    piVar4 = DAT_1803bcb00;
    if (*DAT_1803bcb00 + 1 != 0) {
      *DAT_1803bcb00 = *DAT_1803bcb00 + 1;
    }
    *(int **)(plVar7[3] + 0x70) = piVar4;
    piVar4 = DAT_1803bcb08;
    if (*DAT_1803bcb08 + 1 != 0) {
      *DAT_1803bcb08 = *DAT_1803bcb08 + 1;
    }
    *(int **)(plVar7[3] + 0x78) = piVar4;
    piVar4 = DAT_1803bcb10;
    if (*DAT_1803bcb10 + 1 != 0) {
      *DAT_1803bcb10 = *DAT_1803bcb10 + 1;
    }
    *(int **)(plVar7[3] + 0x80) = piVar4;
    lVar8 = FUN_1801d3910(param_1);
    plVar14 = plVar7;
    if (lVar8 == 0) {
      FUN_180373340(param_1,&local_res8,DAT_1803bcb18);
      uVar13 = 0xd5;
      lVar8 = local_res8;
    }
    else {
      plVar10 = (longlong *)FUN_18037f0c0();
      uVar13 = 0xd5;
      if (plVar10 != (longlong *)0x0) {
        *(undefined4 *)(plVar5 + 5) = 0xd5;
        plVar9 = (longlong *)FUN_180376770(param_1,lVar8,plVar10);
        if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
          (**(code **)(plVar10[1] + 0x30))(plVar10);
        }
        if (plVar9 != (longlong *)0x0) {
          cVar6 = FUN_180372360(param_1,plVar7,plVar9);
          if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
            (**(code **)(plVar9[1] + 0x30))(plVar9);
          }
          if (cVar6 != '\0') {
            lVar8 = PyUnicode_Join(DAT_1803bc9e0,plVar7);
            if (lVar8 != 0) {
              FUN_180001d80(param_1);
              if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                (**(code **)(plVar7[1] + 0x30))(plVar7);
              }
              if ((int)*plVar1 < 0) {
                return lVar8;
              }
              *plVar1 = *plVar1 + -1;
              if (*plVar1 != 0) {
                return lVar8;
              }
              (**(code **)(plVar1[1] + 0x30))(plVar1);
              return lVar8;
            }
            uVar13 = 0xd6;
          }
        }
      }
      lVar8 = *(longlong *)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
    }
  }
  piVar4 = *(int **)(lVar8 + 0x28);
  if (piVar4 == (int *)0x0) {
    pcVar11 = (code *)FUN_18038e2a0(plVar5,uVar13);
    plVar7 = *(longlong **)(lVar8 + 0x28);
    pcVar12 = pcVar11;
    if (pcVar11 == _Py_NoneStruct_exref) {
      pcVar12 = (code *)0x0;
    }
    else {
LAB_1801d6713:
      pcVar11 = pcVar12;
      if ((pcVar12 != (code *)0x0) && (*(int *)pcVar12 + 1 != 0)) {
        *(int *)pcVar12 = *(int *)pcVar12 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar4 + 6) == plVar5) goto LAB_1801d675c;
    pcVar11 = (code *)FUN_18038e2a0(plVar5,uVar13);
    *(int **)(pcVar11 + 0x10) = piVar4;
    if (*piVar4 + 1 != 0) {
      *piVar4 = *piVar4 + 1;
    }
    plVar7 = *(longlong **)(lVar8 + 0x28);
    pcVar12 = pcVar11;
    if (pcVar11 != _Py_NoneStruct_exref) goto LAB_1801d6713;
    pcVar12 = (code *)0x0;
  }
  *(code **)(lVar8 + 0x28) = pcVar12;
  if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)
     ) {
    (**(code **)(plVar7[1] + 0x30))();
  }
  if (((pcVar11 != (code *)0x0) && (-1 < *(int *)pcVar11)) &&
     (*(longlong *)pcVar11 = *(longlong *)pcVar11 + -1, *(longlong *)pcVar11 == 0)) {
    (**(code **)(*(longlong *)(pcVar11 + 8) + 0x30))(pcVar11);
  }
LAB_1801d675c:
  FUN_18036d0a0(plVar5,&DAT_180391be4,plVar1,plVar14);
  plVar7 = DAT_1803bc888;
  if (plVar5 == DAT_1803bc888) {
    if ((-1 < (int)*DAT_1803bc888) && (*DAT_1803bc888 = *DAT_1803bc888 + -1, *plVar7 == 0)) {
      (**(code **)(plVar7[1] + 0x30))();
    }
    DAT_1803bc888 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if (((plVar14 != (longlong *)0x0) && (-1 < (int)*plVar14)) &&
     (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
    (**(code **)(plVar14[1] + 0x30))(plVar14);
  }
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))(plVar1);
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong *)(param_1 + 0x60) = lVar8;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return 0;
}
