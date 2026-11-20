/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x180161530 */

code * FUN_180161530(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  longlong *plVar2;
  int *piVar3;
  longlong *plVar4;
  int iVar5;
  int iVar6;
  undefined8 *puVar7;
  longlong *plVar8;
  code *pcVar9;
  code *pcVar10;
  longlong *plVar11;
  longlong lVar12;
  code *pcVar13;
  longlong lVar14;
  undefined8 uVar15;
  code *pcVar16;
  code *pcVar17;
  longlong *local_res18;
  code *local_res20;

  plVar4 = DAT_1803b7468;
  pcVar17 = (code *)0x0;
  plVar2 = (longlong *)*param_3;
  local_res18 = (longlong *)0x0;
  pcVar16 = (code *)0x0;
  local_res20 = DAT_180391d28;
  if (DAT_1803b7468 == (longlong *)0x0) {
LAB_18016159b:
    DAT_1803b7468 = (longlong *)FUN_18036cd10(param_1,DAT_1803b7f00,DAT_1803ca750,0x20);
  }
  else if ((1 < *DAT_1803b7468) || (DAT_1803b7468[2] != 0)) {
    if ((DAT_1803b7468 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b7468 && (*DAT_1803b7468 = *DAT_1803b7468 + -1, *plVar4 == 0)))) {
      (**(code **)(plVar4[1] + 0x30))(plVar4);
    }
    goto LAB_18016159b;
  }
  plVar4 = DAT_1803b7468;
  FUN_180001d00(param_1,DAT_1803b7468);
  lVar12 = *(longlong *)(DAT_1803ca748 + 0x20);
  if (*(char *)(lVar12 + 10) == '\0') {
    puVar7 = (undefined8 *)FUN_1800015a0(DAT_1803ca748,DAT_1803b75e0);
    pcVar9 = pcVar17;
    if (puVar7 != (undefined8 *)0x0) {
      pcVar9 = (code *)*puVar7;
    }
LAB_18016169f:
    if (pcVar9 == (code *)0x0) goto LAB_1801616a4;
LAB_1801616ed:
    uVar15 = 0x6f;
    *(undefined4 *)(plVar4 + 5) = 0x6f;
    pcVar9 = (code *)FUN_180376210(param_1);
    if (pcVar9 == (code *)0x0) {
      pcVar10 = *(code **)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      pcVar9 = (code *)0x0;
      pcVar17 = pcVar16;
    }
    else {
      uVar15 = 0x70;
      *(undefined4 *)(plVar4 + 5) = 0x70;
      plVar8 = (longlong *)
               FUN_18037e6c0(param_1,plVar2,DAT_1803b75b0,*(undefined8 *)(DAT_1803b75e8 + 0x18));
      if (plVar8 == (longlong *)0x0) {
        pcVar10 = *(code **)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
        pcVar17 = pcVar16;
      }
      else {
        pcVar10 = (code *)FUN_18037f0c0();
        local_res18 = plVar8;
        if (pcVar10 == (code *)0x0) {
          pcVar10 = *(code **)(param_1 + 0x60);
          uVar15 = 0x71;
          *(undefined8 *)(param_1 + 0x60) = 0;
          pcVar17 = pcVar16;
        }
        else {
          local_res20 = _Py_NoneStruct_exref;
          if ((-1 < *(int *)pcVar10) &&
             (*(longlong *)pcVar10 = *(longlong *)pcVar10 + -1, *(longlong *)pcVar10 == 0)) {
            (**(code **)(*(longlong *)(pcVar10 + 8) + 0x30))(pcVar10);
          }
          pcVar13 = pcVar9;
          if (pcVar10 == local_res20) goto LAB_180161b71;
          plVar11 = (longlong *)FUN_18037f0c0(local_res20,plVar8,DAT_1803b75f0);
          uVar15 = 0x72;
          if (plVar11 == (longlong *)0x0) {
            pcVar10 = *(code **)(param_1 + 0x60);
            *(undefined8 *)(param_1 + 0x60) = 0;
            pcVar17 = pcVar16;
          }
          else {
            *(undefined4 *)(plVar4 + 5) = 0x72;
            local_res20 = (code *)FUN_18037e6c0(param_1,plVar11,DAT_1803b75f8,
                                                *(undefined8 *)(DAT_1803b7600 + 0x18));
            if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
              (**(code **)(plVar11[1] + 0x30))(plVar11);
            }
            pcVar17 = local_res20;
            if (local_res20 == (code *)0x0) {
              pcVar10 = *(code **)(param_1 + 0x60);
              *(undefined8 *)(param_1 + 0x60) = 0;
              pcVar17 = pcVar16;
            }
            else {
              iVar5 = FUN_180001510(local_res20);
              if (iVar5 == -1) {
                uVar15 = 0x73;
              }
              else {
                if (iVar5 == 0) {
LAB_180161b71:
                  FUN_180001d80(param_1);
                  iVar5 = *(int *)pcVar13 + 1;
                  iVar6 = *(int *)pcVar13;
                  if (iVar5 != 0) {
                    *(int *)pcVar13 = iVar5;
                    iVar6 = iVar5;
                  }
                  if ((-1 < iVar6) &&
                     (*(longlong *)pcVar13 = *(longlong *)pcVar13 + -1, *(longlong *)pcVar13 == 0))
                  {
                    (**(code **)(*(longlong *)(pcVar13 + 8) + 0x30))(pcVar13);
                  }
                  if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                    (**(code **)(plVar8[1] + 0x30))(plVar8);
                  }
                  if (((pcVar17 != (code *)0x0) && (-1 < *(int *)pcVar17)) &&
                     (*(longlong *)pcVar17 = *(longlong *)pcVar17 + -1, *(longlong *)pcVar17 == 0))
                  {
                    (**(code **)(*(longlong *)(pcVar17 + 8) + 0x30))(pcVar17);
                  }
                  if ((int)*plVar2 < 0) {
                    return pcVar13;
                  }
                  *plVar2 = *plVar2 + -1;
                  if (*plVar2 != 0) {
                    return pcVar13;
                  }
                  (**(code **)(plVar2[1] + 0x30))(plVar2);
                  return pcVar13;
                }
                lVar12 = FUN_180375520(param_1,pcVar9,DAT_1803b7598);
                if (lVar12 != 0) {
                  lVar14 = *(longlong *)(param_1 + 0x10);
                  plVar11 = *(longlong **)(lVar14 + 0x41da8);
                  if (plVar11 == (longlong *)0x0) {
                    plVar11 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,3);
                  }
                  else {
                    *(longlong *)(lVar14 + 0x41da8) = plVar11[3];
                    *(int *)(lVar14 + 0x41e40) = *(int *)(lVar14 + 0x41e40) + -1;
                    *plVar11 = 1;
                  }
                  plVar1 = plVar11 + -2;
                  plVar11[3] = 0;
                  plVar11[4] = 0;
                  plVar11[5] = 0;
                  lVar14 = _PyThreadState_GetCurrent();
                  lVar14 = *(longlong *)(*(longlong *)(lVar14 + 0x10) + 0xd0);
                  puVar7 = *(undefined8 **)(lVar14 + 8);
                  *puVar7 = plVar1;
                  plVar11[-1] = plVar11[-1] & 3;
                  plVar11[-1] = plVar11[-1] | (ulonglong)puVar7;
                  *plVar1 = lVar14;
                  *(longlong **)(lVar14 + 8) = plVar1;
                  plVar11[3] = lVar12;
                  piVar3 = DAT_1803b7608;
                  if (*DAT_1803b7608 + 1 != 0) {
                    *DAT_1803b7608 = *DAT_1803b7608 + 1;
                  }
                  plVar11[4] = (longlong)piVar3;
                  lVar12 = FUN_180375520(param_1,local_res20,DAT_1803b7598);
                  if (lVar12 == 0) {
                    pcVar10 = *(code **)(param_1 + 0x60);
                    uVar15 = 0x74;
                    *(undefined8 *)(param_1 + 0x60) = 0;
                    if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                      (**(code **)(plVar11[1] + 0x30))(plVar11);
                    }
                    goto LAB_1801619aa;
                  }
                  plVar11[5] = lVar12;
                  pcVar13 = (code *)PyUnicode_Join(DAT_1803b7598,plVar11);
                  if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                    (**(code **)(plVar11[1] + 0x30))(plVar11);
                  }
                  if (pcVar13 != (code *)0x0) {
                    if ((-1 < *(int *)pcVar9) &&
                       (*(longlong *)pcVar9 = *(longlong *)pcVar9 + -1, *(longlong *)pcVar9 == 0)) {
                      (**(code **)(*(longlong *)(pcVar9 + 8) + 0x30))();
                    }
                    goto LAB_180161b71;
                  }
                }
                uVar15 = 0x74;
              }
              pcVar10 = *(code **)(param_1 + 0x60);
              *(undefined8 *)(param_1 + 0x60) = 0;
            }
          }
        }
      }
    }
  }
  else {
    iVar5 = *(int *)(lVar12 + 0xc);
    if (iVar5 == 0) {
      iVar5 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
      *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar5 + 1;
      *(int *)(lVar12 + 0xc) = iVar5;
    }
    if (iVar5 != DAT_1803a9984) {
      DAT_1803a9984 = iVar5;
      DAT_1803b74e8 = FUN_1803707d0(lVar12,DAT_1803b75e0,*(undefined8 *)(DAT_1803b75e0 + 0x18));
    }
    if (-1 < DAT_1803b74e8) {
      lVar14 = (1L << (*(byte *)(lVar12 + 9) & 0x3f)) + lVar12;
      if (*(longlong *)(lVar14 + 0x28 + DAT_1803b74e8 * 0x10) != 0) goto LAB_1801616ed;
      DAT_1803b74e8 = FUN_1803707d0(lVar12,DAT_1803b75e0,*(undefined8 *)(DAT_1803b75e0 + 0x18));
      if (-1 < DAT_1803b74e8) {
        pcVar9 = *(code **)(lVar14 + 0x28 + DAT_1803b74e8 * 0x10);
        goto LAB_18016169f;
      }
    }
LAB_1801616a4:
    plVar8 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803b75e0);
    if ((plVar8 != (longlong *)0x0) && (*plVar8 != 0)) goto LAB_1801616ed;
    FUN_180373340(param_1,&local_res20,DAT_1803b75e0);
    uVar15 = 0x6f;
    pcVar10 = local_res20;
    pcVar9 = pcVar17;
  }
LAB_1801619aa:
  piVar3 = *(int **)(pcVar10 + 0x28);
  if (piVar3 == (int *)0x0) {
    pcVar16 = (code *)FUN_18038e2a0(plVar4,uVar15);
    plVar8 = *(longlong **)(pcVar10 + 0x28);
    pcVar13 = pcVar16;
    if (pcVar16 == _Py_NoneStruct_exref) {
      pcVar13 = (code *)0x0;
    }
    else {
LAB_180161a04:
      pcVar16 = pcVar13;
      if ((pcVar13 != (code *)0x0) && (*(int *)pcVar13 + 1 != 0)) {
        *(int *)pcVar13 = *(int *)pcVar13 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar3 + 6) == plVar4) goto LAB_180161a4d;
    pcVar16 = (code *)FUN_18038e2a0(plVar4,uVar15);
    *(int **)(pcVar16 + 0x10) = piVar3;
    if (*piVar3 + 1 != 0) {
      *piVar3 = *piVar3 + 1;
    }
    plVar8 = *(longlong **)(pcVar10 + 0x28);
    pcVar13 = pcVar16;
    if (pcVar16 != _Py_NoneStruct_exref) goto LAB_180161a04;
    pcVar13 = (code *)0x0;
  }
  *(code **)(pcVar10 + 0x28) = pcVar13;
  if (((plVar8 != (longlong *)0x0) && (-1 < (int)*plVar8)) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)
     ) {
    (**(code **)(plVar8[1] + 0x30))();
  }
  if (((pcVar16 != (code *)0x0) && (-1 < *(int *)pcVar16)) &&
     (*(longlong *)pcVar16 = *(longlong *)pcVar16 + -1, *(longlong *)pcVar16 == 0)) {
    (**(code **)(*(longlong *)(pcVar16 + 8) + 0x30))(pcVar16);
  }
LAB_180161a4d:
  FUN_18036d0a0(plVar4,&DAT_180391c1c,plVar2,pcVar9,local_res18,pcVar17);
  plVar8 = DAT_1803b7468;
  if (plVar4 == DAT_1803b7468) {
    if ((-1 < (int)*DAT_1803b7468) && (*DAT_1803b7468 = *DAT_1803b7468 + -1, *plVar8 == 0)) {
      (**(code **)(plVar8[1] + 0x30))();
    }
    DAT_1803b7468 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if (((pcVar9 != (code *)0x0) && (-1 < *(int *)pcVar9)) &&
     (*(longlong *)pcVar9 = *(longlong *)pcVar9 + -1, *(longlong *)pcVar9 == 0)) {
    (**(code **)(*(longlong *)(pcVar9 + 8) + 0x30))(pcVar9);
  }
  if (((local_res18 != (longlong *)0x0) && (-1 < (int)*local_res18)) &&
     (*local_res18 = *local_res18 + -1, *local_res18 == 0)) {
    (**(code **)(local_res18[1] + 0x30))(local_res18);
  }
  if (((pcVar17 != (code *)0x0) && (-1 < *(int *)pcVar17)) &&
     (*(longlong *)pcVar17 = *(longlong *)pcVar17 + -1, *(longlong *)pcVar17 == 0)) {
    (**(code **)(*(longlong *)(pcVar17 + 8) + 0x30))(pcVar17);
  }
  if ((-1 < (int)*plVar2) && (*plVar2 = *plVar2 + -1, *plVar2 == 0)) {
    (**(code **)(plVar2[1] + 0x30))(plVar2);
  }
  plVar2 = *(longlong **)(param_1 + 0x60);
  *(code **)(param_1 + 0x60) = pcVar10;
  if (((plVar2 != (longlong *)0x0) && (-1 < (int)*plVar2)) && (*plVar2 = *plVar2 + -1, *plVar2 == 0)
     ) {
    (**(code **)(plVar2[1] + 0x30))();
  }
  return (code *)0x0;
}
