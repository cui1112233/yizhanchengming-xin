/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x180125650 */

longlong FUN_180125650(longlong param_1,longlong param_2,undefined8 *param_3)

{
  longlong *plVar1;
  int *piVar2;
  longlong *plVar3;
  longlong *plVar4;
  longlong lVar5;
  code *pcVar6;
  code *pcVar7;
  uint uVar8;
  longlong lVar9;
  longlong lVar10;
  longlong local_38;
  longlong *local_30;

  plVar3 = DAT_1803b3a30;
  plVar1 = (longlong *)*param_3;
  if (DAT_1803b3a30 == (longlong *)0x0) {
LAB_18012569a:
    DAT_1803b3a30 = (longlong *)FUN_18036cd10(param_1,DAT_1803b3d50,DAT_1803ca778,0x18);
  }
  else if ((1 < *DAT_1803b3a30) || (DAT_1803b3a30[2] != 0)) {
    if ((DAT_1803b3a30 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b3a30 && (*DAT_1803b3a30 = *DAT_1803b3a30 + -1, *plVar3 == 0)))) {
      (**(code **)(plVar3[1] + 0x30))(plVar3);
    }
    goto LAB_18012569a;
  }
  plVar3 = DAT_1803b3a30;
  FUN_180001d00(param_1,DAT_1803b3a30);
  lVar5 = *(longlong *)(*(longlong *)(param_2 + 0xe8) + 0x10);
  lVar9 = DAT_1803b3fb8;
  if (lVar5 == 0) {
LAB_1801256ef:
    lVar5 = lVar9 + 0x28;
    if (((byte)*(undefined4 *)(lVar9 + 0x20) & 0x60) != 0x60) {
      lVar5 = *(longlong *)(lVar9 + 0x30);
    }
    plVar4 = (longlong *)
             PyUnicode_FromFormat
                       ("cannot access free variable \'%s\' where it is not associated with a value in enclosing scope"
                        ,lVar5);
    lVar5 = FUN_180376770(param_1,*(undefined8 *)PyExc_NameError_exref,plVar4);
    if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
      (**(code **)(plVar4[1] + 0x30))(plVar4);
    }
    FUN_18000d1d0(param_1,lVar5);
  }
  else {
    lVar9 = *(longlong *)(lVar5 + 8);
    pcVar6 = *(code **)(lVar9 + 0x90);
    lVar10 = DAT_1803b52f0;
    if (pcVar6 == (code *)0x0) {
      pcVar6 = *(code **)(lVar9 + 0x40);
      uVar8 = *(uint *)(DAT_1803b52f0 + 0x20) & 0x60;
      if (pcVar6 != (code *)0x0) {
        lVar10 = DAT_1803b52f0 + 0x28;
        if (uVar8 != 0x60) {
          lVar10 = *(longlong *)(DAT_1803b52f0 + 0x30);
        }
        goto LAB_18012578d;
      }
      lVar5 = DAT_1803b52f0 + 0x28;
      if (uVar8 != 0x60) {
        lVar5 = *(longlong *)(DAT_1803b52f0 + 0x30);
      }
      PyErr_Format(*(undefined8 *)PyExc_AttributeError_exref,"\'%s\' object has no attribute \'%s\'"
                   ,*(undefined8 *)(lVar9 + 0x18),lVar5);
    }
    else {
LAB_18012578d:
      plVar4 = (longlong *)(*pcVar6)(lVar5,lVar10);
      if (plVar4 != (longlong *)0x0) {
        lVar5 = *(longlong *)(*(longlong *)(param_2 + 0xe0) + 0x10);
        if (lVar5 == 0) {
          lVar9 = DAT_1803b42a0;
          if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, lVar9 = DAT_1803b42a0, *plVar4 == 0))
          {
            (**(code **)(plVar4[1] + 0x30))(plVar4);
            lVar9 = DAT_1803b42a0;
          }
          goto LAB_1801256ef;
        }
        *(undefined4 *)(plVar3 + 5) = 0x594;
        local_38 = lVar5;
        local_30 = plVar1;
        lVar5 = FUN_180377750(param_1,plVar4,&local_38);
        if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
          (**(code **)(plVar4[1] + 0x30))(plVar4);
        }
        if (lVar5 != 0) {
          FUN_180001d80(param_1);
          if ((int)*plVar1 < 0) {
            return lVar5;
          }
          *plVar1 = *plVar1 + -1;
          if (*plVar1 != 0) {
            return lVar5;
          }
          (**(code **)(plVar1[1] + 0x30))(plVar1);
          return lVar5;
        }
      }
    }
    lVar5 = *(longlong *)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
  }
  piVar2 = *(int **)(lVar5 + 0x28);
  if (piVar2 == (int *)0x0) {
    pcVar6 = (code *)FUN_18038e2a0(plVar3,0x594);
    plVar4 = *(longlong **)(lVar5 + 0x28);
    pcVar7 = pcVar6;
    if (pcVar6 == _Py_NoneStruct_exref) {
      pcVar7 = (code *)0x0;
    }
    else {
LAB_1801258d5:
      pcVar6 = pcVar7;
      if ((pcVar7 != (code *)0x0) && (*(int *)pcVar7 + 1 != 0)) {
        *(int *)pcVar7 = *(int *)pcVar7 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar2 + 6) == plVar3) goto LAB_18012591e;
    pcVar6 = (code *)FUN_18038e2a0(plVar3,0x594);
    *(int **)(pcVar6 + 0x10) = piVar2;
    if (*piVar2 + 1 != 0) {
      *piVar2 = *piVar2 + 1;
    }
    plVar4 = *(longlong **)(lVar5 + 0x28);
    pcVar7 = pcVar6;
    if (pcVar6 != _Py_NoneStruct_exref) goto LAB_1801258d5;
    pcVar7 = (code *)0x0;
  }
  *(code **)(lVar5 + 0x28) = pcVar7;
  if (((plVar4 != (longlong *)0x0) && (-1 < (int)*plVar4)) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)
     ) {
    (**(code **)(plVar4[1] + 0x30))();
  }
  if (((pcVar6 != (code *)0x0) && (-1 < *(int *)pcVar6)) &&
     (*(longlong *)pcVar6 = *(longlong *)pcVar6 + -1, *(longlong *)pcVar6 == 0)) {
    (**(code **)(*(longlong *)(pcVar6 + 8) + 0x30))(pcVar6);
  }
LAB_18012591e:
  FUN_18036d0a0(plVar3,&DAT_18039493c,plVar1,*(undefined8 *)(param_2 + 0xe8),
                *(undefined8 *)(param_2 + 0xe0));
  plVar4 = DAT_1803b3a30;
  if (plVar3 == DAT_1803b3a30) {
    if ((-1 < (int)*DAT_1803b3a30) && (*DAT_1803b3a30 = *DAT_1803b3a30 + -1, *plVar4 == 0)) {
      (**(code **)(plVar4[1] + 0x30))();
    }
    DAT_1803b3a30 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))(plVar1);
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong *)(param_1 + 0x60) = lVar5;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return 0;
}
