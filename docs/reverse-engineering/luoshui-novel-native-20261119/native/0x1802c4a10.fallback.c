/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x1802c4a10 */

code * FUN_1802c4a10(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  longlong *plVar2;
  int *piVar3;
  longlong *plVar4;
  char cVar5;
  int iVar6;
  undefined8 *puVar7;
  longlong *plVar8;
  code *pcVar9;
  longlong *plVar10;
  longlong *plVar11;
  longlong lVar12;
  longlong *plVar13;
  code *pcVar14;
  code *pcVar15;
  int iVar16;
  longlong lVar17;
  longlong lVar18;
  undefined8 uVar19;
  code *local_res20;
  code *local_90;
  code *local_88;
  longlong *local_70;
  longlong *local_68;
  code *local_60;
  code *local_58;
  longlong *local_50;
  longlong *local_48;
  
  plVar4 = DAT_1803c4c10;
  plVar1 = (longlong *)param_3[1];
  pcVar15 = (code *)0x0;
  plVar2 = (longlong *)*param_3;
  local_70 = DAT_180391e98;
  local_60 = (code *)0x0;
  local_res20 = (code *)0x0;
  local_58 = (code *)0x0;
  local_90 = (code *)0x0;
  local_88 = (code *)0x0;
  if (DAT_1803c4c10 == (longlong *)0x0) {
LAB_1802c4aa1:
    DAT_1803c4c10 = (longlong *)FUN_18036cd10(param_1,DAT_1803c4f00,DAT_1803ca468,0x30);
  }
  else if ((1 < *DAT_1803c4c10) || (DAT_1803c4c10[2] != 0)) {
    if ((DAT_1803c4c10 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803c4c10 && (*DAT_1803c4c10 = *DAT_1803c4c10 + -1, *plVar4 == 0)))) {
      (**(code **)(plVar4[1] + 0x30))();
    }
    goto LAB_1802c4aa1;
  }
  plVar4 = DAT_1803c4c10;
  FUN_180001d00(param_1,DAT_1803c4c10);
  lVar12 = *(longlong *)(DAT_1803ca470 + 0x20);
  if (*(char *)(lVar12 + 10) == '\0') {
    puVar7 = (undefined8 *)FUN_1800015a0(DAT_1803ca470,DAT_1803c4d10);
    pcVar9 = pcVar15;
    if (puVar7 != (undefined8 *)0x0) {
      pcVar9 = (code *)*puVar7;
    }
LAB_1802c4bae:
    if (pcVar9 == (code *)0x0) goto LAB_1802c4bb3;
LAB_1802c4bfb:
    plVar8 = (longlong *)FUN_18037f0c0();
    uVar19 = 0x44;
    if (plVar8 == (longlong *)0x0) {
      plVar8 = *(longlong **)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      goto LAB_1802c581f;
    }
    *(undefined4 *)(plVar4 + 5) = 0x44;
    local_50 = plVar2;
    local_48 = plVar8;
    pcVar9 = (code *)FUN_180377750(param_1,pcVar9,&local_50);
    if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
      (**(code **)(plVar8[1] + 0x30))(plVar8);
    }
    if (pcVar9 == (code *)0x0) {
      plVar8 = *(longlong **)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      goto LAB_1802c581f;
    }
    local_60 = pcVar9;
    if (pcVar9 != _Py_NoneStruct_exref) {
      if (*(int *)pcVar9 + 1 != 0) {
        *(int *)pcVar9 = *(int *)pcVar9 + 1;
      }
      plVar8 = (longlong *)FUN_18014b8e0(param_1,pcVar9);
      plVar10 = (longlong *)FUN_180001b80(param_1,plVar2,plVar8);
      if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
        (**(code **)(plVar8[1] + 0x30))(plVar8);
      }
      local_res20 = pcVar9;
      if (plVar10 != (longlong *)0x0) {
        plVar8 = (longlong *)FUN_18037f0c0();
        if (plVar8 == (longlong *)0x0) {
          plVar8 = *(longlong **)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
            (**(code **)(plVar10[1] + 0x30))(plVar10);
            uVar19 = 0x47;
            local_90 = (code *)0x0;
            goto LAB_1802c5822;
          }
          goto LAB_1802c4cf4;
        }
        plVar11 = (longlong *)FUN_180381110(plVar10,plVar8);
        if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
          (**(code **)(plVar10[1] + 0x30))(plVar10);
        }
        if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
          (**(code **)(plVar8[1] + 0x30))(plVar8);
        }
        if (plVar11 == (longlong *)0x0) goto LAB_1802c4cec;
        plVar8 = (longlong *)FUN_18037f0c0();
        if (plVar8 == (longlong *)0x0) {
          plVar8 = *(longlong **)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
            (**(code **)(plVar11[1] + 0x30))(plVar11);
          }
          uVar19 = 0x47;
          local_90 = (code *)0x0;
          goto LAB_1802c5822;
        }
        lVar12 = FUN_180373860(plVar8);
        plVar10 = (longlong *)(param_1 + 0x60);
        if ((lVar12 < 0) && (*plVar10 != 0)) {
          plVar13 = (longlong *)0x0;
        }
        else {
          plVar13 = (longlong *)PyLong_FromSsize_t(lVar12);
        }
        if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
          (**(code **)(plVar8[1] + 0x30))(plVar8);
        }
        if (plVar13 == (longlong *)0x0) {
          plVar8 = (longlong *)*plVar10;
          *plVar10 = 0;
          if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
            (**(code **)(plVar11[1] + 0x30))(plVar11);
          }
          uVar19 = 0x47;
          goto LAB_1802c5822;
        }
        plVar8 = (longlong *)FUN_180380b80(pcVar9,plVar13);
        if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
          (**(code **)(plVar13[1] + 0x30))(plVar13);
        }
        if (plVar8 == (longlong *)0x0) {
          plVar8 = *(longlong **)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
LAB_1802c4ea3:
          if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
            (**(code **)(plVar11[1] + 0x30))(plVar11);
          }
LAB_1802c4ec0:
          uVar19 = 0x47;
          goto LAB_1802c5822;
        }
        plVar10 = (longlong *)FUN_18014b800(param_1,plVar8,_Py_NoneStruct_exref);
        if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
          (**(code **)(plVar8[1] + 0x30))(plVar8);
        }
        plVar8 = (longlong *)FUN_180001b80(param_1,plVar2,plVar10);
        if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
          (**(code **)(plVar10[1] + 0x30))(plVar10);
        }
        if (plVar8 == (longlong *)0x0) {
          plVar8 = *(longlong **)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          goto LAB_1802c4ea3;
        }
        pcVar15 = (code *)FUN_180381110(plVar11,plVar8);
        if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
          (**(code **)(plVar11[1] + 0x30))(plVar11);
        }
        if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
          (**(code **)(plVar8[1] + 0x30))(plVar8);
        }
        if (pcVar15 == (code *)0x0) {
          plVar8 = *(longlong **)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          goto LAB_1802c4ec0;
        }
LAB_1802c5718:
        FUN_180001d80(param_1);
        pcVar14 = local_58;
LAB_1802c5a2e:
        if ((-1 < *(int *)pcVar9) &&
           (*(longlong *)pcVar9 = *(longlong *)pcVar9 + -1, *(longlong *)pcVar9 == 0)) {
          (**(code **)(*(longlong *)(pcVar9 + 8) + 0x30))(pcVar9);
        }
        if (((local_res20 != (code *)0x0) && (-1 < *(int *)local_res20)) &&
           (*(longlong *)local_res20 = *(longlong *)local_res20 + -1, *(longlong *)local_res20 == 0)
           ) {
          (**(code **)(*(longlong *)(local_res20 + 8) + 0x30))();
        }
        if (((pcVar14 != (code *)0x0) && (-1 < *(int *)pcVar14)) &&
           (*(longlong *)pcVar14 = *(longlong *)pcVar14 + -1, *(longlong *)pcVar14 == 0)) {
          (**(code **)(*(longlong *)(pcVar14 + 8) + 0x30))(pcVar14);
        }
        if (((local_90 != (code *)0x0) && (-1 < *(int *)local_90)) &&
           (*(longlong *)local_90 = *(longlong *)local_90 + -1, *(longlong *)local_90 == 0)) {
          (**(code **)(*(longlong *)(local_90 + 8) + 0x30))();
        }
        if ((-1 < (int)*plVar2) && (*plVar2 = *plVar2 + -1, *plVar2 == 0)) {
          (**(code **)(plVar2[1] + 0x30))();
        }
        if ((int)*plVar1 < 0) {
          return pcVar15;
        }
        *plVar1 = *plVar1 + -1;
        if (*plVar1 != 0) {
          return pcVar15;
        }
        (**(code **)(plVar1[1] + 0x30))(plVar1);
        return pcVar15;
      }
LAB_1802c4cec:
      plVar8 = *(longlong **)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
LAB_1802c4cf4:
      uVar19 = 0x47;
      local_90 = pcVar15;
      goto LAB_1802c5822;
    }
    lVar12 = *(longlong *)(DAT_1803ca470 + 0x20);
    if (*(char *)(lVar12 + 10) == '\0') {
      plVar8 = (longlong *)FUN_1800015a0(DAT_1803ca470,DAT_1803c4d20);
      if (plVar8 == (longlong *)0x0) {
        lVar17 = 0;
      }
      else {
        lVar17 = *plVar8;
      }
LAB_1802c5084:
      if (lVar17 == 0) goto LAB_1802c5089;
    }
    else {
      iVar6 = *(int *)(lVar12 + 0xc);
      if (iVar6 == 0) {
        iVar6 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
        *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar6 + 1;
        *(int *)(lVar12 + 0xc) = iVar6;
      }
      if (iVar6 != DAT_1803a9ec0) {
        DAT_1803a9ec0 = iVar6;
        DAT_1803c4bb8 = FUN_1803707d0(lVar12,DAT_1803c4d20,*(undefined8 *)(DAT_1803c4d20 + 0x18));
      }
      if (-1 < DAT_1803c4bb8) {
        lVar18 = (1L << (*(byte *)(lVar12 + 9) & 0x3f)) + lVar12;
        lVar17 = *(longlong *)(lVar18 + 0x28 + DAT_1803c4bb8 * 0x10);
        if (lVar17 != 0) goto LAB_1802c50d6;
        DAT_1803c4bb8 = FUN_1803707d0(lVar12,DAT_1803c4d20,*(undefined8 *)(DAT_1803c4d20 + 0x18));
        if (-1 < DAT_1803c4bb8) {
          lVar17 = *(longlong *)(lVar18 + 0x28 + DAT_1803c4bb8 * 0x10);
          goto LAB_1802c5084;
        }
      }
LAB_1802c5089:
      plVar8 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803c4d20);
      if ((plVar8 == (longlong *)0x0) || (lVar17 = *plVar8, lVar17 == 0)) {
        FUN_180373340(param_1,&local_70,DAT_1803c4d20);
        uVar19 = 0x48;
        plVar8 = local_70;
        goto LAB_1802c581f;
      }
    }
LAB_1802c50d6:
    plVar8 = (longlong *)FUN_18037f0c0();
    uVar19 = 0x48;
    if (plVar8 == (longlong *)0x0) {
      plVar8 = *(longlong **)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      goto LAB_1802c581f;
    }
    *(undefined4 *)(plVar4 + 5) = 0x48;
    local_70 = plVar2;
    local_68 = plVar8;
    pcVar14 = (code *)FUN_180377750(param_1,lVar17,&local_70);
    if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
      (**(code **)(plVar8[1] + 0x30))(plVar8);
    }
    if (pcVar14 == (code *)0x0) {
      plVar8 = *(longlong **)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      goto LAB_1802c581f;
    }
    local_58 = pcVar14;
    if (pcVar14 == _Py_NoneStruct_exref) {
      FUN_180001d80(param_1);
      pcVar15 = _Py_NoneStruct_exref;
      if (*(int *)_Py_NoneStruct_exref + 1 != 0) {
        *(int *)_Py_NoneStruct_exref = *(int *)_Py_NoneStruct_exref + 1;
      }
      goto LAB_1802c5a2e;
    }
    if (*(code **)(*(longlong *)(pcVar14 + 8) + 0xd8) == (code *)0x0) {
      iVar6 = PySequence_Check(pcVar14);
      if (iVar6 == 0) {
        PyErr_Format(*(undefined8 *)PyExc_TypeError_exref,"cannot unpack non-iterable %s object",
                     *(undefined8 *)(*(longlong *)(pcVar14 + 8) + 0x18));
      }
      else {
        plVar10 = (longlong *)FUN_180001220(PySeqIter_Type_exref);
        plVar10[2] = 0;
        if (*(int *)pcVar14 + 1 != 0) {
          *(int *)pcVar14 = *(int *)pcVar14 + 1;
        }
        plVar10[3] = (longlong)pcVar14;
        plVar8 = plVar10 + -2;
        lVar12 = _PyThreadState_GetCurrent();
        lVar12 = *(longlong *)(*(longlong *)(lVar12 + 0x10) + 0xd0);
        plVar11 = *(longlong **)(lVar12 + 8);
        *plVar11 = (longlong)plVar8;
        plVar10[-1] = plVar10[-1] & 3;
        plVar10[-1] = plVar10[-1] | (ulonglong)plVar11;
        *plVar8 = lVar12;
        *(longlong **)(lVar12 + 8) = plVar8;
        if (plVar10 != (longlong *)0x0) goto LAB_1802c524f;
      }
LAB_1802c5805:
      plVar8 = *(longlong **)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
LAB_1802c5812:
      uVar19 = 0x4a;
      goto LAB_1802c581f;
    }
    plVar10 = (longlong *)(**(code **)(*(longlong *)(pcVar14 + 8) + 0xd8))();
    if (plVar10 == (longlong *)0x0) goto LAB_1802c5805;
    lVar12 = *(longlong *)(plVar10[1] + 0xe0);
    if ((lVar12 == 0) || (lVar12 == DAT_1803ca310)) {
      PyErr_Format(*(undefined8 *)PyExc_TypeError_exref,
                   "iter() returned non-iterator of type \'%s\'",*(undefined8 *)(plVar10[1] + 0x18))
      ;
      if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
        (**(code **)(plVar10[1] + 0x30))(plVar10);
      }
      goto LAB_1802c5805;
    }
LAB_1802c524f:
    local_res20 = (code *)(**(code **)(plVar10[1] + 0xe0))(plVar10);
    if (local_res20 == (code *)0x0) {
      plVar8 = *(longlong **)(param_1 + 0x60);
      local_res20 = pcVar15;
      local_88 = pcVar15;
      if ((plVar8 != (longlong *)0x0) && (lVar12 = plVar8[1], lVar12 != 0)) {
        lVar17 = *(longlong *)(lVar12 + 8);
        if ((*(uint *)(*(longlong *)(lVar12 + 8) + 0xa8) & 0x40000000) == 0) {
          lVar17 = lVar12;
        }
        if ((lVar17 != *(longlong *)PyExc_StopIteration_exref) &&
           (((-1 < *(int *)(*(longlong *)(lVar17 + 8) + 0xa8) ||
             ((*(uint *)(lVar17 + 0xa8) & 0x40000000) == 0)) ||
            (cVar5 = FUN_180382640(), cVar5 != '\x01')))) {
          *(undefined8 *)(param_1 + 0x60) = 0;
          goto LAB_1802c5787;
        }
        *(undefined8 *)(param_1 + 0x60) = 0;
        if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
          (**(code **)(plVar8[1] + 0x30))(plVar8);
        }
      }
      plVar11 = (longlong *)
                PyUnicode_FromFormat("not enough values to unpack (expected %d, got %d)",2,0);
      plVar8 = (longlong *)FUN_180376770(param_1,*(undefined8 *)PyExc_ValueError_exref,plVar11);
      iVar6 = (int)*plVar11;
LAB_1802c576f:
      if ((-1 < iVar6) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
        (**(code **)(plVar11[1] + 0x30))(plVar11);
      }
LAB_1802c5787:
      if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
        (**(code **)(plVar10[1] + 0x30))(plVar10);
      }
      if (((local_res20 != (code *)0x0) && (-1 < *(int *)local_res20)) &&
         (*(longlong *)local_res20 = *(longlong *)local_res20 + -1, *(longlong *)local_res20 == 0))
      {
        (**(code **)(*(longlong *)(local_res20 + 8) + 0x30))(local_res20);
      }
      if (((local_88 != (code *)0x0) && (-1 < *(int *)local_88)) &&
         (*(longlong *)local_88 = *(longlong *)local_88 + -1, *(longlong *)local_88 == 0)) {
        (**(code **)(*(longlong *)(local_88 + 8) + 0x30))(local_88);
      }
      goto LAB_1802c5812;
    }
    local_90 = (code *)(**(code **)(plVar10[1] + 0xe0))(plVar10);
    if (local_90 == (code *)0x0) {
      plVar8 = *(longlong **)(param_1 + 0x60);
      if ((plVar8 != (longlong *)0x0) && (lVar12 = plVar8[1], lVar12 != 0)) {
        lVar17 = *(longlong *)(lVar12 + 8);
        if ((*(uint *)(*(longlong *)(lVar12 + 8) + 0xa8) & 0x40000000) == 0) {
          lVar17 = lVar12;
        }
        if ((lVar17 != *(longlong *)PyExc_StopIteration_exref) &&
           (((-1 < *(int *)(*(longlong *)(lVar17 + 8) + 0xa8) ||
             ((*(uint *)(lVar17 + 0xa8) & 0x40000000) == 0)) ||
            (cVar5 = FUN_180382640(), cVar5 != '\x01')))) {
          *(undefined8 *)(param_1 + 0x60) = 0;
          goto LAB_1802c5787;
        }
        *(undefined8 *)(param_1 + 0x60) = 0;
        if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
          (**(code **)(plVar8[1] + 0x30))(plVar8);
        }
      }
      plVar11 = (longlong *)
                PyUnicode_FromFormat("not enough values to unpack (expected %d, got %d)",2,1);
      plVar8 = (longlong *)FUN_180376770(param_1,*(undefined8 *)PyExc_ValueError_exref,plVar11);
      if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
        (**(code **)(plVar11[1] + 0x30))(plVar11);
      }
      local_88 = (code *)0x0;
      goto LAB_1802c5787;
    }
    plVar8 = (longlong *)(**(code **)(plVar10[1] + 0xe0))(plVar10);
    local_88 = local_90;
    if (plVar8 != (longlong *)0x0) {
      if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
        (**(code **)(plVar8[1] + 0x30))(plVar8);
      }
      plVar11 = (longlong *)PyUnicode_FromFormat("too many values to unpack (expected %d)",2);
      plVar8 = (longlong *)FUN_180376770(param_1,*(undefined8 *)PyExc_ValueError_exref,plVar11);
      iVar6 = (int)*plVar11;
      goto LAB_1802c576f;
    }
    plVar8 = *(longlong **)(param_1 + 0x60);
    if ((plVar8 != (longlong *)0x0) && (lVar12 = plVar8[1], lVar12 != 0)) {
      lVar17 = *(longlong *)(lVar12 + 8);
      if ((*(uint *)(*(longlong *)(lVar12 + 8) + 0xa8) & 0x40000000) == 0) {
        lVar17 = lVar12;
      }
      if (lVar17 != *(longlong *)PyExc_StopIteration_exref) {
        if ((*(int *)(*(longlong *)(lVar17 + 8) + 0xa8) < 0) &&
           ((*(uint *)(lVar17 + 0xa8) & 0x40000000) != 0)) {
          cVar5 = FUN_180382640(lVar17,*(longlong *)PyExc_StopIteration_exref);
          *(undefined8 *)(param_1 + 0x60) = 0;
          if (cVar5 == '\x01') goto LAB_1802c54ab;
        }
        else {
          *(undefined8 *)(param_1 + 0x60) = 0;
        }
        goto LAB_1802c5787;
      }
      *(undefined8 *)(param_1 + 0x60) = 0;
LAB_1802c54ab:
      if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
        (**(code **)(plVar8[1] + 0x30))(plVar8);
      }
    }
    if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
      (**(code **)(plVar10[1] + 0x30))(plVar10);
    }
    iVar6 = *(int *)local_res20 + 1;
    iVar16 = *(int *)local_res20;
    if (iVar6 != 0) {
      *(int *)local_res20 = iVar6;
      iVar16 = iVar6;
    }
    if ((-1 < iVar16) &&
       (*(longlong *)local_res20 = *(longlong *)local_res20 + -1, *(longlong *)local_res20 == 0)) {
      (**(code **)(*(longlong *)(local_res20 + 8) + 0x30))(local_res20);
    }
    iVar6 = *(int *)local_90 + 1;
    iVar16 = *(int *)local_90;
    if (iVar6 != 0) {
      *(int *)local_90 = iVar6;
      iVar16 = iVar6;
    }
    if ((-1 < iVar16) &&
       (*(longlong *)local_90 = *(longlong *)local_90 + -1, *(longlong *)local_90 == 0)) {
      (**(code **)(*(longlong *)(local_90 + 8) + 0x30))(local_90);
    }
    plVar8 = (longlong *)FUN_18014b8e0(param_1,local_res20);
    plVar10 = (longlong *)FUN_180001b80(param_1,plVar2,plVar8);
    if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
      (**(code **)(plVar8[1] + 0x30))(plVar8);
    }
    if (plVar10 == (longlong *)0x0) {
LAB_1802c5577:
      plVar8 = *(longlong **)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
LAB_1802c5583:
      uVar19 = 0x4b;
    }
    else {
      plVar8 = (longlong *)FUN_18037f0c0();
      if (plVar8 != (longlong *)0x0) {
        plVar11 = (longlong *)FUN_180381110(plVar10,plVar8);
        if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
          (**(code **)(plVar10[1] + 0x30))(plVar10);
        }
        if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
          (**(code **)(plVar8[1] + 0x30))(plVar8);
        }
        if (plVar11 != (longlong *)0x0) {
          plVar8 = (longlong *)FUN_18014b800(param_1,local_90,_Py_NoneStruct_exref);
          plVar10 = (longlong *)FUN_180001b80(param_1,plVar2,plVar8);
          if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
            (**(code **)(plVar8[1] + 0x30))(plVar8);
          }
          if (plVar10 == (longlong *)0x0) {
            plVar8 = *(longlong **)(param_1 + 0x60);
            *(undefined8 *)(param_1 + 0x60) = 0;
            if (((int)*plVar11 < 0) || (*plVar11 = *plVar11 + -1, *plVar11 != 0))
            goto LAB_1802c5583;
            (**(code **)(plVar11[1] + 0x30))(plVar11);
            uVar19 = 0x4b;
            goto LAB_1802c5822;
          }
          pcVar15 = (code *)FUN_180381110(plVar11,plVar10);
          if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
            (**(code **)(plVar11[1] + 0x30))(plVar11);
          }
          if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
            (**(code **)(plVar10[1] + 0x30))(plVar10);
          }
          if (pcVar15 != (code *)0x0) goto LAB_1802c5718;
        }
        goto LAB_1802c5577;
      }
      plVar8 = *(longlong **)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      if (((int)*plVar10 < 0) || (*plVar10 = *plVar10 + -1, *plVar10 != 0)) goto LAB_1802c5583;
      (**(code **)(plVar10[1] + 0x30))(plVar10);
      uVar19 = 0x4b;
    }
  }
  else {
    iVar6 = *(int *)(lVar12 + 0xc);
    if (iVar6 == 0) {
      iVar6 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
      *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar6 + 1;
      *(int *)(lVar12 + 0xc) = iVar6;
    }
    if (iVar6 != DAT_1803a9ebc) {
      DAT_1803a9ebc = iVar6;
      DAT_1803c4b58 = FUN_1803707d0(lVar12,DAT_1803c4d10,*(undefined8 *)(DAT_1803c4d10 + 0x18));
    }
    if (-1 < DAT_1803c4b58) {
      lVar17 = (1L << (*(byte *)(lVar12 + 9) & 0x3f)) + lVar12;
      pcVar9 = *(code **)(lVar17 + 0x28 + DAT_1803c4b58 * 0x10);
      if (pcVar9 != (code *)0x0) goto LAB_1802c4bfb;
      DAT_1803c4b58 = FUN_1803707d0(lVar12,DAT_1803c4d10,*(undefined8 *)(DAT_1803c4d10 + 0x18));
      if (-1 < DAT_1803c4b58) {
        pcVar9 = *(code **)(lVar17 + 0x28 + DAT_1803c4b58 * 0x10);
        goto LAB_1802c4bae;
      }
    }
LAB_1802c4bb3:
    puVar7 = (undefined8 *)FUN_1800015a0(DAT_1803ab778,DAT_1803c4d10);
    if ((puVar7 != (undefined8 *)0x0) && (pcVar9 = (code *)*puVar7, pcVar9 != (code *)0x0))
    goto LAB_1802c4bfb;
    FUN_180373340(param_1,&local_70,DAT_1803c4d10);
    uVar19 = 0x44;
    plVar8 = local_70;
LAB_1802c581f:
    local_90 = (code *)0x0;
    local_res20 = (code *)0x0;
  }
LAB_1802c5822:
  piVar3 = (int *)plVar8[5];
  if (piVar3 == (int *)0x0) {
    pcVar15 = (code *)FUN_18038e2a0(plVar4,uVar19);
    plVar10 = (longlong *)plVar8[5];
    pcVar9 = pcVar15;
    if (pcVar15 == _Py_NoneStruct_exref) {
      pcVar9 = (code *)0x0;
    }
    else {
LAB_1802c5883:
      pcVar15 = pcVar9;
      if ((pcVar9 != (code *)0x0) && (*(int *)pcVar9 + 1 != 0)) {
        *(int *)pcVar9 = *(int *)pcVar9 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar3 + 6) == plVar4) goto LAB_1802c58cc;
    pcVar15 = (code *)FUN_18038e2a0(plVar4,uVar19);
    *(int **)(pcVar15 + 0x10) = piVar3;
    if (*piVar3 + 1 != 0) {
      *piVar3 = *piVar3 + 1;
    }
    plVar10 = (longlong *)plVar8[5];
    pcVar9 = pcVar15;
    if (pcVar15 != _Py_NoneStruct_exref) goto LAB_1802c5883;
    pcVar9 = (code *)0x0;
  }
  plVar8[5] = (longlong)pcVar9;
  if (((plVar10 != (longlong *)0x0) && (-1 < (int)*plVar10)) &&
     (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
    (**(code **)(plVar10[1] + 0x30))();
  }
  if (((pcVar15 != (code *)0x0) && (-1 < *(int *)pcVar15)) &&
     (*(longlong *)pcVar15 = *(longlong *)pcVar15 + -1, *(longlong *)pcVar15 == 0)) {
    (**(code **)(*(longlong *)(pcVar15 + 8) + 0x30))(pcVar15);
  }
LAB_1802c58cc:
  pcVar15 = local_58;
  FUN_18036d0a0(plVar4,"oooooo",plVar2,plVar1,local_60,local_res20,local_58,local_90);
  plVar10 = DAT_1803c4c10;
  if (plVar4 == DAT_1803c4c10) {
    if ((-1 < (int)*DAT_1803c4c10) && (*DAT_1803c4c10 = *DAT_1803c4c10 + -1, *plVar10 == 0)) {
      (**(code **)(plVar10[1] + 0x30))();
    }
    DAT_1803c4c10 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if (((local_60 != (code *)0x0) && (-1 < *(int *)local_60)) &&
     (*(longlong *)local_60 = *(longlong *)local_60 + -1, *(longlong *)local_60 == 0)) {
    (**(code **)(*(longlong *)(local_60 + 8) + 0x30))();
  }
  if (((local_res20 != (code *)0x0) && (-1 < *(int *)local_res20)) &&
     (*(longlong *)local_res20 = *(longlong *)local_res20 + -1, *(longlong *)local_res20 == 0)) {
    (**(code **)(*(longlong *)(local_res20 + 8) + 0x30))(local_res20);
  }
  if (((pcVar15 != (code *)0x0) && (-1 < *(int *)pcVar15)) &&
     (*(longlong *)pcVar15 = *(longlong *)pcVar15 + -1, *(longlong *)pcVar15 == 0)) {
    (**(code **)(*(longlong *)(pcVar15 + 8) + 0x30))(pcVar15);
  }
  if (((local_90 != (code *)0x0) && (-1 < *(int *)local_90)) &&
     (*(longlong *)local_90 = *(longlong *)local_90 + -1, *(longlong *)local_90 == 0)) {
    (**(code **)(*(longlong *)(local_90 + 8) + 0x30))(local_90);
  }
  if ((-1 < (int)*plVar2) && (*plVar2 = *plVar2 + -1, *plVar2 == 0)) {
    (**(code **)(plVar2[1] + 0x30))(plVar2);
  }
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))(plVar1);
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong **)(param_1 + 0x60) = plVar8;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return (code *)0x0;
}

