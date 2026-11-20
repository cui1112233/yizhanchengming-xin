/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x180160f00 */

/* WARNING: Function: __security_check_cookie replaced with injection: security_check_cookie */

longlong FUN_180160f00(longlong param_1,longlong param_2)

{
  undefined8 *puVar1;
  int *piVar2;
  longlong *plVar3;
  longlong *plVar4;
  longlong lVar5;
  longlong *plVar6;
  longlong *plVar7;
  code *pcVar8;
  code *pcVar9;
  uint uVar10;
  longlong lVar11;
  longlong lVar12;
  undefined8 uVar13;
  undefined1 auStack_b8 [32];
  undefined8 local_98;
  undefined8 local_90;
  undefined8 local_88;
  longlong local_70;
  longlong *local_68;
  longlong *local_60;
  longlong *local_58;
  int *local_50;
  ulonglong local_48;

  plVar3 = DAT_1803b7380;
  local_48 = DAT_1803a6040 ^ (ulonglong)auStack_b8;
  local_70 = DAT_180391d28;
  if (DAT_1803b7380 == (longlong *)0x0) {
LAB_180160f67:
    DAT_1803b7380 = (longlong *)FUN_18036cd10(param_1,DAT_1803b73a0,DAT_1803ca750,0x28);
  }
  else if ((1 < *DAT_1803b7380) || (DAT_1803b7380[2] != 0)) {
    if ((DAT_1803b7380 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b7380 && (*DAT_1803b7380 = *DAT_1803b7380 + -1, *plVar3 == 0)))) {
      (**(code **)(plVar3[1] + 0x30))(plVar3);
    }
    goto LAB_180160f67;
  }
  plVar3 = DAT_1803b7380;
  FUN_180001d00(param_1,DAT_1803b7380);
  lVar5 = *(longlong *)(*(longlong *)(param_2 + 0xf8) + 0x10);
  lVar11 = DAT_1803b75a8;
  if (lVar5 == 0) {
LAB_180160fd2:
    lVar5 = lVar11 + 0x28;
    if (((byte)*(undefined4 *)(lVar11 + 0x20) & 0x60) != 0x60) {
      lVar5 = *(longlong *)(lVar11 + 0x30);
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
    lVar11 = *(longlong *)(lVar5 + 8);
    pcVar9 = *(code **)(lVar11 + 0x90);
    lVar12 = DAT_1803b75b0;
    if (pcVar9 == (code *)0x0) {
      pcVar9 = *(code **)(lVar11 + 0x40);
      uVar10 = *(uint *)(DAT_1803b75b0 + 0x20) & 0x60;
      if (pcVar9 != (code *)0x0) {
        lVar12 = DAT_1803b75b0 + 0x28;
        if (uVar10 != 0x60) {
          lVar12 = *(longlong *)(DAT_1803b75b0 + 0x30);
        }
        goto LAB_180161070;
      }
      lVar5 = DAT_1803b75b0 + 0x28;
      if (uVar10 != 0x60) {
        lVar5 = *(longlong *)(DAT_1803b75b0 + 0x30);
      }
      PyErr_Format(*(undefined8 *)PyExc_AttributeError_exref,"\'%s\' object has no attribute \'%s\'"
                   ,*(undefined8 *)(lVar11 + 0x18),lVar5);
    }
    else {
LAB_180161070:
      plVar4 = (longlong *)(*pcVar9)(lVar5,lVar12);
      if (plVar4 != (longlong *)0x0) {
        if (*(longlong *)(*(longlong *)(param_2 + 0xe0) + 0x10) == 0) {
          lVar11 = DAT_1803b75b8;
          if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, lVar11 = DAT_1803b75b8, *plVar4 == 0))
          {
            (**(code **)(plVar4[1] + 0x30))(plVar4);
            lVar11 = DAT_1803b75b8;
          }
          goto LAB_180160fd2;
        }
        *(undefined4 *)(plVar3 + 5) = 0x6a;
        plVar6 = (longlong *)FUN_180376770(param_1,plVar4);
        if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
          (**(code **)(plVar4[1] + 0x30))(plVar4);
        }
        if (plVar6 != (longlong *)0x0) {
          plVar4 = (longlong *)FUN_18037f0c0();
          if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
            (**(code **)(plVar6[1] + 0x30))(plVar6);
          }
          if (plVar4 != (longlong *)0x0) {
            piVar2 = *(int **)(*(longlong *)(param_2 + 0xf0) + 0x10);
            if (piVar2 == (int *)0x0) {
              if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
                (**(code **)(plVar4[1] + 0x30))(plVar4);
              }
              FUN_180371050(param_1,&local_70,DAT_1803b75c8);
              lVar5 = local_70;
              FUN_18000d1d0(param_1,local_70);
            }
            else {
              lVar5 = *(longlong *)(param_1 + 0x10);
              plVar6 = *(longlong **)(lVar5 + 0x41d98);
              if (plVar6 == (longlong *)0x0) {
                plVar6 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,1);
              }
              else {
                *(longlong *)(lVar5 + 0x41d98) = plVar6[3];
                *(int *)(lVar5 + 0x41e38) = *(int *)(lVar5 + 0x41e38) + -1;
                *plVar6 = 1;
              }
              plVar6[3] = 0;
              plVar7 = plVar6 + -2;
              lVar5 = _PyThreadState_GetCurrent();
              lVar5 = *(longlong *)(*(longlong *)(lVar5 + 0x10) + 0xd0);
              puVar1 = *(undefined8 **)(lVar5 + 8);
              *puVar1 = plVar7;
              plVar6[-1] = plVar6[-1] & 3;
              plVar6[-1] = plVar6[-1] | (ulonglong)puVar1;
              *plVar7 = lVar5;
              *(longlong **)(lVar5 + 8) = plVar7;
              if (*piVar2 + 1 != 0) {
                *piVar2 = *piVar2 + 1;
              }
              plVar6[3] = (longlong)piVar2;
              uVar13 = DAT_1803b75d0;
              lVar5 = *(longlong *)(*(longlong *)(param_2 + 0x100) + 0x10);
              if (lVar5 == 0) {
                if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
                  (**(code **)(plVar4[1] + 0x30))(plVar4);
                }
                uVar13 = DAT_1803b75d0;
                if ((-1 < (int)*plVar6) &&
                   (*plVar6 = *plVar6 + -1, uVar13 = DAT_1803b75d0, *plVar6 == 0)) {
                  (**(code **)(plVar6[1] + 0x30))(plVar6);
                  uVar13 = DAT_1803b75d0;
                }
              }
              else {
                plVar7 = (longlong *)_PyDict_NewPresized(1);
                PyDict_SetItem(plVar7,uVar13,lVar5);
                piVar2 = *(int **)(*(longlong *)(param_2 + 0xe8) + 0x10);
                if (piVar2 != (int *)0x0) {
                  if (*piVar2 + 1 != 0) {
                    *piVar2 = *piVar2 + 1;
                  }
                  local_68 = plVar4;
                  local_60 = plVar6;
                  local_58 = plVar7;
                  local_50 = piVar2;
                  lVar5 = FUN_180098520(param_1,&local_68);
                  if (lVar5 != 0) {
                    FUN_180001d80(param_1);
                    return lVar5;
                  }
                  goto LAB_180161352;
                }
                if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
                  (**(code **)(plVar4[1] + 0x30))(plVar4);
                }
                if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                  (**(code **)(plVar6[1] + 0x30))(plVar6);
                }
                uVar13 = DAT_1803b75d8;
                if ((-1 < (int)*plVar7) &&
                   (*plVar7 = *plVar7 + -1, uVar13 = DAT_1803b75d8, *plVar7 == 0)) {
                  (**(code **)(plVar7[1] + 0x30))(plVar7);
                  uVar13 = DAT_1803b75d8;
                }
              }
              FUN_180371050(param_1,&local_70,uVar13);
              lVar5 = local_70;
              FUN_18000d1d0(param_1,local_70);
            }
            goto LAB_18016135a;
          }
        }
      }
    }
LAB_180161352:
    lVar5 = *(longlong *)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
  }
LAB_18016135a:
  pcVar9 = (code *)0x0;
  piVar2 = *(int **)(lVar5 + 0x28);
  if (piVar2 == (int *)0x0) {
    pcVar8 = (code *)FUN_18038e2a0(plVar3,0x6a);
    plVar4 = *(longlong **)(lVar5 + 0x28);
    if (pcVar8 != _Py_NoneStruct_exref) {
LAB_1801613bc:
      pcVar9 = pcVar8;
      pcVar8 = pcVar9;
      if ((pcVar9 != (code *)0x0) && (*(int *)pcVar9 + 1 != 0)) {
        *(int *)pcVar9 = *(int *)pcVar9 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar2 + 6) == plVar3) goto LAB_180161405;
    pcVar8 = (code *)FUN_18038e2a0(plVar3,0x6a);
    *(int **)(pcVar8 + 0x10) = piVar2;
    if (*piVar2 + 1 != 0) {
      *piVar2 = *piVar2 + 1;
    }
    plVar4 = *(longlong **)(lVar5 + 0x28);
    if (pcVar8 != _Py_NoneStruct_exref) goto LAB_1801613bc;
  }
  *(code **)(lVar5 + 0x28) = pcVar9;
  if (((plVar4 != (longlong *)0x0) && (-1 < (int)*plVar4)) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)
     ) {
    (**(code **)(plVar4[1] + 0x30))();
  }
  if (((pcVar8 != (code *)0x0) && (-1 < *(int *)pcVar8)) &&
     (*(longlong *)pcVar8 = *(longlong *)pcVar8 + -1, *(longlong *)pcVar8 == 0)) {
    (**(code **)(*(longlong *)(pcVar8 + 8) + 0x30))(pcVar8);
  }
LAB_180161405:
  local_88 = *(undefined8 *)(param_2 + 0xe8);
  local_90 = *(undefined8 *)(param_2 + 0x100);
  local_98 = *(undefined8 *)(param_2 + 0xf0);
  FUN_18036d0a0(plVar3,"ccccc",*(undefined8 *)(param_2 + 0xf8),*(undefined8 *)(param_2 + 0xe0));
  plVar4 = DAT_1803b7380;
  if (plVar3 == DAT_1803b7380) {
    if ((-1 < (int)*DAT_1803b7380) && (*DAT_1803b7380 = *DAT_1803b7380 + -1, *plVar4 == 0)) {
      (**(code **)(plVar4[1] + 0x30))();
    }
    DAT_1803b7380 = (longlong *)0x0;
  }
  lVar11 = **(longlong **)(param_1 + 0x38);
  plVar3 = *(longlong **)(lVar11 + 0x30);
  **(longlong **)(param_1 + 0x38) = *(longlong *)(lVar11 + 8);
  *(undefined4 *)(plVar3 + 8) = 0xffffffff;
  plVar4 = (longlong *)plVar3[2];
  if (((plVar4 != (longlong *)0x0) && (plVar3[2] = 0, -1 < (int)*plVar4)) &&
     (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
    (**(code **)(plVar4[1] + 0x30))();
  }
  if ((-1 < (int)*plVar3) && (*plVar3 = *plVar3 + -1, *plVar3 == 0)) {
    (**(code **)(plVar3[1] + 0x30))(plVar3);
  }
  plVar3[10] = 0;
  plVar3 = *(longlong **)(param_1 + 0x60);
  *(longlong *)(param_1 + 0x60) = lVar5;
  if (((plVar3 != (longlong *)0x0) && (-1 < (int)*plVar3)) && (*plVar3 = *plVar3 + -1, *plVar3 == 0)
     ) {
    (**(code **)(plVar3[1] + 0x30))();
  }
  return 0;
}
